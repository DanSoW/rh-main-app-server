package model

/* Структура основной информации в таблице */
type HeaderInfoModel struct {
	Title              string   `json:"title"`
	AddressItem        []string `json:"address_item"`
	TimeDelivery       string   `json:"time_delivery"`
	PaymentVariant     []string `json:"payment_variant"`
	PropertyItem       []string `json:"property_item"`
	CommunicateVariant []string `json:"communicate_variant"`
}

/* Структура идентификатора ячейки таблицы */
type IndexCellModel struct {
	Pos    string `json:"pos"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Value  string `json:"string"`
}
