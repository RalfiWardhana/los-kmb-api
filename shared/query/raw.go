package query

import "fmt"

func ScanInstallmentAmountWgOff(idNumber, name, birthDate, surgate string) string {

	var filter3Days string = `
	AND DATE_FORMAT(DtmUpd ,'%Y-%m-%d') >= DATE_FORMAT(SUBDATE(NOW(), INTERVAL 3 DAY) ,'%Y-%m-%d')
	`

	wgOff := fmt.Sprintf(`
	SELECT IDNumber, LegalName, BirthDate, SurgateMotherName, SUM(InstallmentAmount) AS InstallmentAmount,
	SUM(NTF) AS NTF FROM eform_inquiry WHERE ((status_process = 4 AND status_decision = 1 AND GoLiveDate IS NULL) 
	OR status_process <> 4) AND IDNumber = '%s' AND LegalName = '%s' AND BirthDate = '%s' 
	AND SurgateMotherName = '%s' %s`, idNumber, name, birthDate, surgate, filter3Days)

	return wgOff
}

func ScanInstallmentAmountKmbOff(idNumber, name, birthDate, surgate string) string {

	var filter3Days string = `
	AND DATE_FORMAT(a.DtmUpd ,'%Y-%m-%d') >= DATE_FORMAT(SUBDATE(NOW(), INTERVAL 3 DAY) ,'%Y-%m-%d')
	`

	kmbOff := fmt.Sprintf(`
	SELECT IDNumber, LegalName, BirthDate, SurgateMotherName, 
	SUM(InstallmentAmount) AS InstallmentAmount, SUM(NTF) AS NTF FROM data_inquiry a 
	LEFT JOIN final_inquiry b ON a.ProspectID = b.ProspectID WHERE 
	((b.final_approval = 1 AND a.GoLiveDate IS NULL) OR b.final_approval IS NULL)
	AND IDNumber = '%s' AND LegalName = '%s' AND BirthDate = '%s' AND SurgateMotherName = '%s'
	%s`, idNumber, name, birthDate, surgate, filter3Days)

	return kmbOff
}

func ScanInstallmentAmountKmobOFF(idNumber, name, birthDate, surgate string) string {

	kmob := fmt.Sprintf(`
	SELECT
	SUM(ta.InstallmentAmount) AS InstallmentAmount,SUM(ta.NTF) AS NTF FROM trx_apk ta WITH (nolock)
		LEFT JOIN confins_agreements ca WITH (nolock) ON ca.ProspectID = ta.ProspectID 
		INNER JOIN customer_personal cp WITH (nolock) ON cp.ProspectID = ta.ProspectID 
		INNER JOIN trx_status ts WITH (nolock) ON ts.ProspectID = ta.ProspectID 
		INNER JOIN trx_master tm WITH (nolock) ON tm.ProspectID = ta.ProspectID
		WHERE CAST(ta.created_at AS date) >= DATEADD(day, -3, CAST(GETDATE() AS date)) 
		AND cp.IDNumber = '%s' AND cp.LegalName = '%s' AND cp.BirthDate = '%s' AND cp.SurgateMotherName = '%s' AND tm.lob = 'KMOB' AND ((ts.decision = 'APR' AND ContractStatus = '0') OR ts.decision = 'CPR')
	`, idNumber, name, birthDate, surgate)

	return kmob
}

func ScanInstallmentAmountWgONL(idNumber, name, birthDate, surgate string) string {
	wgOnl := fmt.Sprintf(`
	SELECT
	SUM(ta.InstallmentAmount) AS InstallmentAmount,SUM(ta.NTF) AS NTF FROM trx_apk ta WITH (nolock)
		LEFT JOIN confins_agreements ca WITH (nolock) ON ca.ProspectID = ta.ProspectID 
		INNER JOIN customer_personal cp WITH (nolock) ON cp.ProspectID = ta.ProspectID 
		INNER JOIN trx_status ts WITH (nolock) ON ts.ProspectID = ta.ProspectID 
		INNER JOIN trx_master tm WITH (nolock) ON tm.ProspectID = ta.ProspectID
		WHERE CAST(ta.created_at AS date) >= DATEADD(day, -3, CAST(GETDATE() AS date)) 
		AND cp.IDNumber = '%s' AND cp.LegalName = '%s' AND cp.BirthDate = '%s' AND cp.SurgateMotherName = '%s' 
		AND tm.lob = 'WG' AND tm.channel = 'ONL' AND tm.transaction_type IN ('PRODUCT_LIMIT','USE_LIMIT') AND ((ts.decision = 'APR' AND ContractStatus = '0') OR (ts.decision = 'APR' AND ContractStatus IS NULL) OR ts.decision = 'CPR' OR ts.decision = 'PAS' OR (ts.decision IS NULL AND ts.activity = 'PRGS'))
	`, idNumber, name, birthDate, surgate)
	return wgOnl
}
