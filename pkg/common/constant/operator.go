package constant

const (
	OperatorIsat      = "Isat"
	OperatorTelkomsel = "Telkomsel"
	OperatorXL        = "XL"
	OperatorAxis      = "Axis"
	OperatorThree     = "Three"
	OperatorSmartfren = "Smartfren"
	OperatorUnknown   = "unknown"
)

var OperatorByPrefix = map[string]string{
	"0811": OperatorTelkomsel,
	"0813": OperatorTelkomsel,
	"0821": OperatorTelkomsel,
	"0823": OperatorTelkomsel,
	"0851": OperatorTelkomsel,
	"0853": OperatorTelkomsel,
	"0852": OperatorTelkomsel,
	"0822": OperatorTelkomsel,
	"0812": OperatorTelkomsel,

	"0814": OperatorIsat,
	"0816": OperatorIsat,
	"0855": OperatorIsat,
	"0858": OperatorIsat,
	"0857": OperatorIsat,
	"0856": OperatorIsat,
	"0898": OperatorIsat,
	"0815": OperatorIsat,

	"0817": OperatorXL,
	"0819": OperatorXL,
	"0859": OperatorXL,
	"0877": OperatorXL,
	"0878": OperatorXL,
	"0818": OperatorXL,

	"0831": OperatorAxis,
	"0833": OperatorAxis,
	"0838": OperatorAxis,

	"0895": OperatorThree,
	"0896": OperatorThree,
	"0899": OperatorThree,

	"0881": OperatorSmartfren,
	"0882": OperatorSmartfren,
	"0888": OperatorSmartfren,
	"0889": OperatorSmartfren,
}
