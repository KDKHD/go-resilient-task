package taskproperties

type IKafkaProperties interface {
	GetBootstrapServers() string
	GetKafkaTopicsNamespace() string
	GetKafkaConsumerGroupId() string
}

type KafkaProperties struct {
	bootstrapServers     string
	kafkaTopicsNamespace string
	kafkaConsumerGroupId string
}

type KafkaPropertiesConfigOption func(*KafkaProperties)

func WithBootstrapServers(bootstrapServers string) KafkaPropertiesConfigOption {
	return func(config *KafkaProperties) {
		config.bootstrapServers = bootstrapServers
	}
}

func WithKafkaTopicsNamespace(kafkaTopicsNamespace string) KafkaPropertiesConfigOption {
	return func(config *KafkaProperties) {
		config.kafkaTopicsNamespace = kafkaTopicsNamespace
	}
}

func WithKafkaConsumerGroupId(kafkaConsumerGroupId string) KafkaPropertiesConfigOption {
	return func(config *KafkaProperties) {
		config.kafkaConsumerGroupId = kafkaConsumerGroupId
	}
}

func NewKafkaProperties(options ...KafkaPropertiesConfigOption) *KafkaProperties {
	config := &KafkaProperties{
		bootstrapServers:     "localhost:9092",
		kafkaTopicsNamespace: "",
		kafkaConsumerGroupId: "default-consumer-group",
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func (kp KafkaProperties) GetBootstrapServers() string {
	return kp.bootstrapServers
}

func (kp KafkaProperties) GetKafkaTopicsNamespace() string {
	return kp.kafkaTopicsNamespace
}

func (kp KafkaProperties) GetKafkaConsumerGroupId() string {
	return kp.kafkaConsumerGroupId
}
