package settlementfx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/onecombine/onecombine-msg-validator/src/partners"
)

type SettlementFX struct {
	Pair     string `json:"pair"`
	Value    string `json:"value"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

const (
	API_LIST_SETTLEMENT_FX_PATH = "/v1/switching/fx"
)

type SettlementFXService struct {
	baseUrl    string
	fxStore    *partners.MemoryStore
	fxConsumer SettlementFXConsumer

	cls chan string
	wg  *sync.WaitGroup
}

func NewSettlementFXService(baseUrl string, kConfig *partners.KafkaConfig) *SettlementFXService {
	store := partners.NewMemoryStore()

	consumer := NewKafkaSettlementFXConsumer(store, kConfig)
	var wg sync.WaitGroup

	service := &SettlementFXService{
		baseUrl:    baseUrl,
		fxStore:    store,
		fxConsumer: consumer,
		wg:         &wg,
	}

	if err := service.refreshSettlementFX(); err != nil {
		fmt.Printf("Error refresh the settlement fx, error: %v\n", err)
	}

	service.cls = service.fxConsumer.Subscribe(service.wg)

	return service
}

func (s *SettlementFXService) GetFXStore() *partners.MemoryStore {
	return s.fxStore
}

func (s *SettlementFXService) WaitForCompletion() {
	s.wg.Wait()
}

func (s *SettlementFXService) GracefullyShutdown() {
	go func() {
		s.cls <- "done"
	}()
}

func (s *SettlementFXService) refreshSettlementFX() error {
	response, err := http.Get(fmt.Sprintf("%s%s", s.baseUrl, API_LIST_SETTLEMENT_FX_PATH))

	if err != nil {
		return nil
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var fxs []*SettlementFX
	json.Unmarshal(responseData, &fxs)

	for _, fx := range fxs {
		s.fxStore.Set(fx.Pair, fx)
	}

	return nil
}
