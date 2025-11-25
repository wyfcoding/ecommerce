package idgen

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"

	"github.com/bwmarrin/snowflake"
)

// Generator defines the ID generator interface.
type Generator interface {
	Generate() int64
}

// SnowflakeGenerator implements Generator using Snowflake algorithm.
type SnowflakeGenerator struct {
	node *snowflake.Node
}

// NewSnowflakeGenerator creates a new SnowflakeGenerator.
func NewSnowflakeGenerator(cfg config.SnowflakeConfig) (*SnowflakeGenerator, error) {
	// Set start time if provided
	if cfg.StartTime != "" {
		st, err := time.Parse("2006-01-02", cfg.StartTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		snowflake.Epoch = st.UnixNano() / 1000000
	}

	node, err := snowflake.NewNode(cfg.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	return &SnowflakeGenerator{
		node: node,
	}, nil
}

// Generate generates a new ID.
func (g *SnowflakeGenerator) Generate() int64 {
	return g.node.Generate().Int64()
}

// Global default generator
var (
	defaultGenerator *SnowflakeGenerator
	once             sync.Once
)

// Init initializes the global default generator.
func Init(cfg config.SnowflakeConfig) error {
	var err error
	once.Do(func() {
		defaultGenerator, err = NewSnowflakeGenerator(cfg)
	})
	return err
}

// GenID generates a unique ID using the default generator.
func GenID() uint64 {
	if defaultGenerator == nil {
		// Fallback initialization with default values if not initialized
		_ = Init(config.SnowflakeConfig{MachineID: 1})
	}
	return uint64(defaultGenerator.Generate())
}

// GenOrderNo generates an order number.
func GenOrderNo() string {
	return fmt.Sprintf("O%d", GenID())
}

// GenPaymentNo generates a payment number.
func GenPaymentNo() string {
	return fmt.Sprintf("P%d", GenID())
}

// GenRefundNo generates a refund number.
func GenRefundNo() string {
	return fmt.Sprintf("R%d", GenID())
}

// GenSPUNo generates a SPU number.
func GenSPUNo() string {
	return fmt.Sprintf("SPU%d", GenID())
}

// GenSKUNo generates a SKU number.
func GenSKUNo() string {
	return fmt.Sprintf("SKU%d", GenID())
}

// GenCouponCode generates a coupon code.
func GenCouponCode() string {
	return fmt.Sprintf("C%d", GenID())
}
