package frpc

import (
	"bytes"
	"fmt"
	"github.com/SianHH/frp-package/pkg/config"
	"github.com/SianHH/frp-package/pkg/config/legacy"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/sets"
)

func LoadClientConfig(data []byte, strict bool) (
	*v1.ClientCommonConfig,
	[]v1.ProxyConfigurer,
	[]v1.VisitorConfigurer,
	bool, error,
) {
	var (
		cliCfg         *v1.ClientCommonConfig
		proxyCfgs      = make([]v1.ProxyConfigurer, 0)
		visitorCfgs    = make([]v1.VisitorConfigurer, 0)
		isLegacyFormat bool
	)

	if config.DetectLegacyINIFormat(data) {
		legacyCommon, legacyProxyCfgs, legacyVisitorCfgs, err := ParseClientConfig(data)
		if err != nil {
			return nil, nil, nil, true, err
		}
		cliCfg = legacy.Convert_ClientCommonConf_To_v1(&legacyCommon)
		for _, c := range legacyProxyCfgs {
			proxyCfgs = append(proxyCfgs, legacy.Convert_ProxyConf_To_v1(c))
		}
		for _, c := range legacyVisitorCfgs {
			visitorCfgs = append(visitorCfgs, legacy.Convert_VisitorConf_To_v1(c))
		}
		isLegacyFormat = true
	} else {
		allCfg := v1.ClientConfig{}
		content, err := config.RenderWithTemplate(data, nil)
		if err != nil {
			return nil, nil, nil, false, err
		}

		if err := config.LoadConfigure(content, &allCfg, strict); err != nil {
			return nil, nil, nil, false, err
		}
		cliCfg = &allCfg.ClientCommonConfig
		for _, c := range allCfg.Proxies {
			proxyCfgs = append(proxyCfgs, c.ProxyConfigurer)
		}
		for _, c := range allCfg.Visitors {
			visitorCfgs = append(visitorCfgs, c.VisitorConfigurer)
		}
	}

	// Filter by start
	if len(cliCfg.Start) > 0 {
		startSet := sets.New(cliCfg.Start...)
		proxyCfgs = lo.Filter(proxyCfgs, func(c v1.ProxyConfigurer, _ int) bool {
			return startSet.Has(c.GetBaseConfig().Name)
		})
		visitorCfgs = lo.Filter(visitorCfgs, func(c v1.VisitorConfigurer, _ int) bool {
			return startSet.Has(c.GetBaseConfig().Name)
		})
	}

	if cliCfg != nil {
		cliCfg.Complete()
	}
	for _, c := range proxyCfgs {
		c.Complete(cliCfg.User)
	}
	for _, c := range visitorCfgs {
		c.Complete(cliCfg)
	}
	return cliCfg, proxyCfgs, visitorCfgs, isLegacyFormat, nil
}

func ParseClientConfig(data []byte) (
	cfg legacy.ClientCommonConf,
	proxyCfgs map[string]legacy.ProxyConf,
	visitorCfgs map[string]legacy.VisitorConf,
	err error,
) {
	var content []byte
	content, err = legacy.RenderContent(data)
	if err != nil {
		return
	}
	configBuffer := bytes.NewBuffer(nil)
	configBuffer.Write(content)

	// Parse common section.
	cfg, err = legacy.UnmarshalClientConfFromIni(content)
	if err != nil {
		return
	}
	if err = cfg.Validate(); err != nil {
		err = fmt.Errorf("parse config error: %v", err)
		return
	}

	// Parse all proxy and visitor configs.
	proxyCfgs, visitorCfgs, err = legacy.LoadAllProxyConfsFromIni(cfg.User, configBuffer.Bytes(), cfg.Start)
	if err != nil {
		return
	}
	return
}
