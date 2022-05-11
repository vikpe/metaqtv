package transform

import (
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/proxy"
	"github.com/vikpe/serverstat/qserver/qtv"
	"metaqtv/geo"
)

type MvdsvWithGeo struct {
	mvdsv.Server
	Geo geo.Info
}
type ProxyWithGeo struct {
	proxy.Proxy
	Geo geo.Info
}
type QtvWithGeo struct {
	qtv.Qtv
	Geo geo.Info
}

func ToMvdsvServers(serversWithGeo []geo.ServerWithGeo) []MvdsvWithGeo {
	mvdsvServers := make([]MvdsvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		mvdsvServers = append(mvdsvServers, MvdsvWithGeo{
			Server: mvdsv.Parse(serverWithGeo.GenericServer),
			Geo:    serverWithGeo.Geo,
		})
	}

	return mvdsvServers
}

func ToProxies(serversWithGeo []geo.ServerWithGeo) []ProxyWithGeo {
	proxies := make([]ProxyWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		proxies = append(proxies, ProxyWithGeo{
			Proxy: proxy.Parse(serverWithGeo.GenericServer),
			Geo:   serverWithGeo.Geo,
		})
	}

	return proxies
}

func ToQtvServers(serversWithGeo []geo.ServerWithGeo) []QtvWithGeo {
	qtvServers := make([]QtvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		qtvServers = append(qtvServers, QtvWithGeo{
			Qtv: qtv.Parse(serverWithGeo.GenericServer),
			Geo: serverWithGeo.Geo,
		})
	}

	return qtvServers
}

func ServerAddressToQtvStreamUrlMap(servers []geo.ServerWithGeo) map[string]string {
	serverToQtv := make(map[string]string, 0)

	for _, server := range servers {
		if "" != server.ExtraInfo.QtvStream.Url {
			serverToQtv[server.Address] = server.ExtraInfo.QtvStream.Url
		}
	}

	return serverToQtv
}

func QtvStreamUrlToServerAddressMap(servers []geo.ServerWithGeo) map[string]string {
	return ReverseStringMap(ServerAddressToQtvStreamUrlMap(servers))
}

func ReverseStringMap(map_ map[string]string) map[string]string {
	reversed := make(map[string]string, 0)
	for key, value := range map_ {
		reversed[value] = key
	}
	return reversed
}
