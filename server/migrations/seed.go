package migrations

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

func SeedCompanies(db *gorm.DB) error {
	companies := []model.Company{
		{Name: "SmartNews", CareersURL: "https://apply.workable.com/smartnews/", ATSType: "workable", ATSSlug: "smartnews", Region: "JP", Location: "Tokyo, Japan", Active: true},
		{Name: "CADDi", CareersURL: "https://apply.workable.com/caddi/", ATSType: "workable", ATSSlug: "caddi", Region: "JP", Location: "Tokyo, Japan", Active: true},
		{Name: "Cookpad", CareersURL: "https://apply.workable.com/cookpad/", ATSType: "workable", ATSSlug: "cookpad", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Mercari", CareersURL: "https://apply.workable.com/mercari/", ATSType: "workable", ATSSlug: "mercari", Region: "JP", Location: "Tokyo, Japan / Remote Japan", Active: true},
		{Name: "PayPay", CareersURL: "https://boards.greenhouse.io/paypay", ATSType: "greenhouse", ATSSlug: "paypay", Region: "JP", Location: "Tokyo, Japan", Active: true},
		{Name: "Paidy", CareersURL: "https://boards.greenhouse.io/paidyinc", ATSType: "greenhouse", ATSSlug: "paidyinc", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Rakuten", CareersURL: "https://global.rakuten.com/corp/careers/", ATSType: "rakuten", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "SmartHR", CareersURL: "https://open.talentio.com/r/1/c/smarthr/pages", ATSType: "talentio", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Cover Corp", CareersURL: "https://cover-corp.com/en/recruit/", ATSType: "cover_corp", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Sansan", CareersURL: "https://open.talentio.com/r/1/c/sansan/pages", ATSType: "talentio", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Money Forward", CareersURL: "https://recruit.moneyforward.com/en", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "GO Inc.", CareersURL: "https://goinc.jp/career/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Hennge", CareersURL: "https://recruit.hennge.com/en/mid-career-ngh/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "LY Corp", CareersURL: "https://www.lycorp.co.jp/en/recruit/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Woven", CareersURL: "https://woven.toyota/en/careers/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Zeals", CareersURL: "https://careers.zeals.ai/en/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "LayerX", CareersURL: "https://jobs.layerx.co.jp/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Wantedly", CareersURL: "https://wantedlyinc.com/en/careers/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Fast Retailing", CareersURL: "https://www.fastretailing.com/careers/en/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Tektome", CareersURL: "https://careers.tektome.com/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
		{Name: "Cogent Labs", CareersURL: "https://www.cogent.co.jp/en/careers/", ATSType: "generic", ATSSlug: "", Region: "JP", Location: "Tokyo, Japan", Active: false},
	}

	for _, c := range companies {
		if err := db.Where(model.Company{Name: c.Name}).FirstOrCreate(&c).Error; err != nil {
			return err
		}
		if c.ATSType != "" && c.ATSSlug != "" {
			board := model.CareerBoard{CompanyID: c.ID, Provider: c.ATSType, BoardIdentifier: c.ATSSlug, CanonicalURL: c.CareersURL, Active: c.Active}
			if err := db.Where("provider = ? AND board_identifier = ?", board.Provider, board.BoardIdentifier).FirstOrCreate(&board).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
