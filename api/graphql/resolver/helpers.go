package resolver

func fail(err error) (bool, error) {
	return false, err
}

func success() (bool, error) {
	return true, nil
}

// func NullDecimalToStringPtr(dec decimal.NullDecimal) *string {
// 	if dec.Valid {
// 		str := dec.Decimal.String()
// 		return &str
// 	}

// 	return nil
// }

// func parseIntID(s string) (int64, error) {
// 	return strconv.ParseInt(s, 10, 64)
// }

// func getIDPartFromSplitID(id string) string {
// 	idParts := strings.Split(id, "_")

// 	if len(idParts) == 2 {
// 		return idParts[1]
// 	}

// 	return id
// }
