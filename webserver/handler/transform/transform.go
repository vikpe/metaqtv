package transform

import (
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/geo"
)

type MvdsvWithGeo struct {
	Type string
	mvdsv.MvdsvExport
	Geo geo.Info
}
type QwfwdWithGeo struct {
	Type string
	qwfwd.Qwfwd
	Geo geo.Info
}
type QtvWithGeo struct {
	Type string
	qtv.Qtv
	Geo geo.Info
}

func ToMvdsvServers(serversWithGeo []geo.ServerWithGeo) []MvdsvWithGeo {
	mvdsvServers := make([]MvdsvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		mvdsvServers = append(mvdsvServers, MvdsvWithGeo{
			Type:        mvdsv.Name,
			MvdsvExport: convert.ToMvdsv(serverWithGeo.GenericServer).Export(),
			Geo:         serverWithGeo.Geo,
		})
	}

	return mvdsvServers
}

func ToQwfwds(serversWithGeo []geo.ServerWithGeo) []QwfwdWithGeo {
	proxies := make([]QwfwdWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		proxies = append(proxies, QwfwdWithGeo{
			Type:  qwfwd.Name,
			Qwfwd: convert.ToQwfwd(serverWithGeo.GenericServer),
			Geo:   serverWithGeo.Geo,
		})
	}

	return proxies
}

func ToQtvServers(serversWithGeo []geo.ServerWithGeo) []QtvWithGeo {
	qtvServers := make([]QtvWithGeo, 0)

	for _, serverWithGeo := range serversWithGeo {
		qtvServers = append(qtvServers, QtvWithGeo{
			Type: qtv.Name,
			Qtv:  convert.ToQtv(serverWithGeo.GenericServer),
			Geo:  serverWithGeo.Geo,
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
