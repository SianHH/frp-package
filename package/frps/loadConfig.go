package frps

import (
	"github.com/SianHH/frp-package/pkg/config"
	"github.com/SianHH/frp-package/pkg/config/legacy"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
)

func LoadServerConfig(data []byte, strict bool) (*v1.ServerConfig, bool, error) {
	var (
		svrCfg         *v1.ServerConfig
		isLegacyFormat bool
	)
	// detect legacy ini format
	if config.DetectLegacyINIFormat(data) {
		content, err := legacy.RenderContent(data)
		if err != nil {
			return nil, true, err
		}
		legacyCfg, err := legacy.UnmarshalServerConfFromIni(content)
		if err != nil {
			return nil, true, err
		}
		svrCfg = legacy.Convert_ServerCommonConf_To_v1(&legacyCfg)
		isLegacyFormat = true
	} else {
		svrCfg = &v1.ServerConfig{}
		content, err := config.RenderWithTemplate(data, nil)
		if err != nil {
			return nil, false, err
		}
		if err := config.LoadConfigure(content, svrCfg, strict); err != nil {
			return nil, false, err
		}
	}
	if svrCfg != nil {
		svrCfg.Complete()
	}
	return svrCfg, isLegacyFormat, nil
}
