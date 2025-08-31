package ftm

import "strings"

// TopicType as enum with fixed set.
type TopicType struct {
	BaseType
	values map[string]string
}

func NewTopicType() *TopicType {
	t := &TopicType{BaseType: BaseType{name: "topic", group: "topics", label: "Topic", matchable: false, maxLength: 64}, values: map[string]string{}}
	for k, v := range map[string]string{
		"crime": "Crime", "crime.fraud": "Fraud", "crime.cyber": "Cybercrime", "crime.fin": "Financial crime", "crime.env": "Environmental violations", "crime.theft": "Theft", "crime.war": "War crimes", "crime.boss": "Criminal leadership", "crime.terror": "Terrorism", "crime.traffick": "Trafficking", "crime.traffick.drug": "Drug trafficking", "crime.traffick.human": "Human trafficking", "forced.labor": "Forced labor", "asset.frozen": "Frozen asset", "wanted": "Wanted", "corp.offshore": "Offshore", "corp.shell": "Shell company", "corp.public": "Public listed company", "corp.disqual": "Disqualified", "gov": "Government", "gov.national": "National government", "gov.state": "State government", "gov.muni": "Municipal government", "gov.soe": "State-owned enterprise", "gov.igo": "Intergovernmental organization", "gov.head": "Head of government or state", "gov.admin": "Civil service", "gov.executive": "Executive branch of government", "gov.legislative": "Legislative branch of government", "gov.judicial": "Judicial branch of government", "gov.security": "Security services", "gov.financial": "Central banking and financial integrity", "gov.religion": "Religious leadership", "fin": "Financial services", "fin.bank": "Bank", "fin.fund": "Fund", "fin.adivsor": "Financial advisor", "mare.detained": "Maritime detention", "mare.shadow": "Shadow fleet", "mare.sts": "Ship-to-ship transfer", "reg.action": "Regulator action", "reg.warn": "Regulator warning", "role.pep": "Politician", "role.pol": "Non-PEP", "role.rca": "Close Associate", "role.judge": "Judge", "role.civil": "Civil servant", "role.diplo": "Diplomat", "role.lawyer": "Lawyer", "role.acct": "Accountant", "role.spy": "Spy", "role.oligarch": "Oligarch", "role.journo": "Journalist", "role.act": "Activist", "role.lobby": "Lobbyist", "pol.party": "Political party", "pol.union": "Union", "rel": "Religion", "mil": "Military", "sanction": "Sanctioned entity", "sanction.linked": "Sanction-linked entity", "sanction.counter": "Counter-sanctioned entity", "export.control": "Export controlled", "export.risk": "Trade risk", "debarment": "Debarred entity", "poi": "Person of interest",
	} {
		t.values[k] = v
	}
	return t
}
func (t *TopicType) Validate(value string) bool {
	_, ok := t.values[strings.ToLower(strings.TrimSpace(value))]
	return ok
}
func (t *TopicType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(text))
	if _, ok := t.values[v]; ok {
		return v, true
	}
	return "", false
}
func (t *TopicType) Caption(value, _ string) string {
	if l, ok := t.values[value]; ok {
		return l
	}
	return value
}
