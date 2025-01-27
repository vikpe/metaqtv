package dataprovider

import (
	"github.com/vikpe/serverstat/qserver"
	"github.com/vikpe/serverstat/qserver/convert"
	"github.com/vikpe/serverstat/qserver/mvdsv"
	"github.com/vikpe/serverstat/qserver/qtv"
	"github.com/vikpe/serverstat/qserver/qwfwd"
	"metaqtv/geodb"
	"metaqtv/scrape"
)

type DataProvider struct {
	scraper *scrape.ServerScraper
	geoDb   geodb.Database
}

func New(scraper *scrape.ServerScraper, geoDb geodb.Database) DataProvider {
	return DataProvider{
		scraper: scraper,
		geoDb:   geoDb,
	}
}

func (dp DataProvider) Generic() []qserver.GenericServer {
	return dp.scraper.Servers()
}

func (dp DataProvider) Mvdsv() []mvdsv.MvdsvExport {
	result := make([]mvdsv.MvdsvExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsMvdsv() && len(server.Clients) > 0 {
			mvdsvExport := convert.ToMvdsvExport(server)

			if mvdsvExport.PlayerSlots.Used > 0 {
				mvdsvExport.Geo = dp.geoDb.GetByAddress(server.Address)
				result = append(result, mvdsvExport)
			}
		}
	}

	return result
}

func (dp DataProvider) Qtv() []qtv.QtvExport {
	result := make([]qtv.QtvExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsQtv() {
			qtvExport := convert.ToQtvExport(server)
			qtvExport.Geo = dp.geoDb.GetByAddress(server.Address)
			result = append(result, qtvExport)
		}
	}

	return result
}

func (dp DataProvider) Qwfwd() []qwfwd.QwfwdExport {
	result := make([]qwfwd.QwfwdExport, 0)

	for _, server := range dp.scraper.Servers() {
		if server.Version.IsQwfwd() {
			qwfwdExport := convert.ToQwfwdExport(server)
			qwfwdExport.Geo = dp.geoDb.GetByAddress(server.Address)
			result = append(result, qwfwdExport)
		}
	}

	return result
}
