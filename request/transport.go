// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/4/15
// 描述：请求拦截
// *****************************************************************************

package request

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

type ProxyRewriteTransport struct {
	Base       http.RoundTripper
	proxyHosts map[string]string
}

func (s *ProxyRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if proxyAddr, ok := s.proxyHosts[req.URL.Host]; ok {

		u, err := url.Parse(proxyAddr)
		if err == nil {
			// 只拼 host + path
			targetPath := req.URL.Host + req.URL.Path

			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = u.Path + targetPath

			// 保留原 query
			req.Host = ""
		}
	}

	return s.Base.RoundTrip(req)
}

// InitProxyRewrite 初始化请求拦截
// 拦截字典
//
//	map[string]string{
//		"api.weixin.qq.com": "https://30.192.1.234:6443/jump/",
//	}
func InitProxyRewrite(hosts map[string]string) {
	http.DefaultTransport = &ProxyRewriteTransport{
		Base: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		proxyHosts: hosts,
	}
}
