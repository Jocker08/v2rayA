package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"log"
)

func StopV2ray() (err error) {
	err = v2ray.StopV2rayService(true)
	if err != nil {
		return
	}
	return
}
func StartV2ray() (err error) {
	if css := configure.GetConnectedServers(); css.Len() == 0 {
		return fmt.Errorf("failed: no server is connected. connect a server instead")
	}
	return v2ray.UpdateV2RayConfig()
}

func IsClassicMode() bool {
	supportLoadBalance := v2ray.CheckObservatorySupported() == nil
	singleOutbound := len(configure.GetOutbounds()) <= 1
	return singleOutbound && !supportLoadBalance
}

func Disconnect(which configure.Which, clearOutbound bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to disconnect: %w", err)
		}
	}()
	lastConnected := configure.GetConnectedServersByOutbound(which.Outbound)
	if clearOutbound {
		err = configure.ClearConnects(which.Outbound)
	} else {
		err = configure.RemoveConnect(which)
	}
	if err != nil {
		return
	}
	//update the v2ray config and restart v2ray
	if v2ray.IsV2RayRunning() || IsClassicMode() {
		defer func() {
			if err != nil && lastConnected != nil && v2ray.IsV2RayRunning() {
				_ = configure.OverwriteConnects(lastConnected)
				_ = v2ray.UpdateV2RayConfig()
			}
		}()
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			return
		}
	}
	return
}

func checkAssetsExist(setting *configure.Setting) error {
	//FIXME: non-fully check
	if setting.RulePortMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !asset.LoyalsoldierSiteDatExists() {
			return newError("GFWList file not exists. Try updating GFWList please")
		}
	}
	return nil
}

func Connect(which *configure.Which) (err error) {
	log.Println("Connect: begin")
	defer log.Println("Connect: done")
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to connect: %w", err)
		}
	}()
	setting := GetSetting()
	if err = checkAssetsExist(setting); err != nil {
		return
	}
	if which == nil {
		return newError("which can not be nil")
	}
	//configure the ip forward
	if setting.IntranetSharing != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IntranetSharing)
		if e != nil {
			log.Println("[Warning]", e)
		}
	}
	//locate server
	currentConnected := configure.GetConnectedServersByOutbound(which.Outbound)
	defer func() {
		// if error occurs, restore the result of connecting
		if err != nil && currentConnected != nil && v2ray.IsV2RayRunning() {
			_ = configure.OverwriteConnects(currentConnected)
			_ = v2ray.UpdateV2RayConfig()
		}
	}()
	//save the result of connecting to database
	supportLoadBalance := v2ray.CheckObservatorySupported() == nil
	if !supportLoadBalance {
		if err = configure.ClearConnects(which.Outbound); err != nil {
			return
		}
	}
	if err = configure.AddConnect(*which); err != nil {
		return
	}
	//update the v2ray config and start/restart v2ray
	if v2ray.IsV2RayRunning() || IsClassicMode() {
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			return
		}
	}
	return
}
