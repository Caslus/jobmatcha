package migrations

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

func SeedCompanies(db *gorm.DB) error {
	companies := []model.Company{
		{Name: "SmartNews", CareersURL: "https://apply.workable.com/smartnews/", ATSType: "workable", ATSSlug: "smartnews", Active: true},
		{Name: "CADDi", CareersURL: "https://apply.workable.com/caddi/", ATSType: "workable", ATSSlug: "caddi", Active: true},
		{Name: "Cookpad", CareersURL: "https://apply.workable.com/cookpad/", ATSType: "workable", ATSSlug: "cookpad", Active: false},
		{Name: "Mercari", CareersURL: "https://apply.workable.com/mercari/", ATSType: "workable", ATSSlug: "mercari", Active: true},
		{Name: "PayPay", CareersURL: "https://boards.greenhouse.io/paypay", ATSType: "greenhouse", ATSSlug: "paypay", Active: true},
		{Name: "Paidy", CareersURL: "https://boards.greenhouse.io/paidyinc", ATSType: "greenhouse", ATSSlug: "paidyinc", Active: false},
		{Name: "A2MAC1", CareersURL: "https://apply.workable.com/a2mac1/", ATSType: "workable", ATSSlug: "a2mac1", Active: true},
		{Name: "AI Robot Association", CareersURL: "https://apply.workable.com/ai-robot-association/", ATSType: "workable", ATSSlug: "ai-robot-association", Active: true},
		{Name: "BoostDraft", CareersURL: "https://apply.workable.com/boostdraft/", ATSType: "workable", ATSSlug: "boostdraft", Active: true},
		{Name: "Clinigen", CareersURL: "https://apply.workable.com/clinigen/", ATSType: "workable", ATSSlug: "clinigen", Active: true},
		{Name: "Eku Energy", CareersURL: "https://apply.workable.com/eku-energy/", ATSType: "workable", ATSSlug: "eku-energy", Active: true},
		{Name: "Eram Talent", CareersURL: "https://apply.workable.com/eramtalent-1/", ATSType: "workable", ATSSlug: "eramtalent-1", Active: true},
		{Name: "ESR Group", CareersURL: "https://apply.workable.com/esr-group/", ATSType: "workable", ATSSlug: "esr-group", Active: true},
		{Name: "Exotec", CareersURL: "https://apply.workable.com/exotec/", ATSType: "workable", ATSSlug: "exotec", Active: true},
		{Name: "Genetec", CareersURL: "https://apply.workable.com/genetec-inc/", ATSType: "workable", ATSSlug: "genetec-inc", Active: true},
		{Name: "GoGlobal", CareersURL: "https://apply.workable.com/goglobal/", ATSType: "workable", ATSSlug: "goglobal", Active: true},
		{Name: "Intellect", CareersURL: "https://apply.workable.com/intellecthq/", ATSType: "workable", ATSSlug: "intellecthq", Active: true},
		{Name: "Komodo Co., Ltd.", CareersURL: "https://apply.workable.com/komodo-co-dot-ltd/", ATSType: "workable", ATSSlug: "komodo-co-dot-ltd", Active: true},
		{Name: "KOMOJU", CareersURL: "https://apply.workable.com/komoju/", ATSType: "workable", ATSSlug: "komoju", Active: true},
		{Name: "Mindrift", CareersURL: "https://apply.workable.com/toloka-ai/", ATSType: "workable", ATSSlug: "toloka-ai", Active: true},
		{Name: "PEOPLECERT", CareersURL: "https://apply.workable.com/peoplecert/", ATSType: "workable", ATSSlug: "peoplecert", Active: true},
		{Name: "PIX4D", CareersURL: "https://apply.workable.com/pix4d/", ATSType: "workable", ATSSlug: "pix4d", Active: true},
		{Name: "Pixelogic Media Partners, LLC", CareersURL: "https://apply.workable.com/pixelogicmedia/", ATSType: "workable", ATSSlug: "pixelogicmedia", Active: true},
		{Name: "Polymer Capital Japan", CareersURL: "https://apply.workable.com/polymer-capital-japan/", ATSType: "workable", ATSSlug: "polymer-capital-japan", Active: true},
		{Name: "PriceLabs", CareersURL: "https://apply.workable.com/pricelabs/", ATSType: "workable", ATSSlug: "pricelabs", Active: true},
		{Name: "Rapyuta Robotics", CareersURL: "https://apply.workable.com/rapyuta-robotics/", ATSType: "workable", ATSSlug: "rapyuta-robotics", Active: true},
		{Name: "SAP Fioneer", CareersURL: "https://apply.workable.com/sap-fioneer/", ATSType: "workable", ATSSlug: "sap-fioneer", Active: true},
		{Name: "Side", CareersURL: "https://apply.workable.com/side/", ATSType: "workable", ATSSlug: "side", Active: true},
		{Name: "SORACOM", CareersURL: "https://apply.workable.com/soracom/", ATSType: "workable", ATSSlug: "soracom", Active: true},
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
