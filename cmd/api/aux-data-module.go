package MaoApi

var (
	AuxDataModuleRegisterName = "aux-data-module"
)

type AuxDataProcessor interface {
	Process(auxData string)
}

type AuxDataModule interface {
	AddProcessor(p *AuxDataProcessor)
}

