package realtimequote

var SupportedTickers = []string{
	"BTC",
	"ETH",
	"DOGE"}

var SupportedProducts = []string{
	"BTC-USD",
	"ETH-USD",
	"DOGE-USD"}

func IsSupported(arr []string, t string) bool {
	for _, v := range arr {
		if v == t {
			return true
		}
	}
	return false
}
