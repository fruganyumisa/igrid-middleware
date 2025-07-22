package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/pkg/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type Publisher struct {
	client mqtt.Client
	logger *zap.Logger
	config config.Config
}

func NewPublisher(cfg config.Config) (*Publisher, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.BrokerURL)
	opts.SetClientID(cfg.MQTT.ClientID)
	opts.SetProtocolVersion(4) // 4 = MQTT 3.1.1
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		zap.L().Error("MQTT connection lost", zap.Error(err))
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		zap.L().Info("MQTT client connected", zap.String("clientID", cfg.MQTT.ClientID))
	})
	opts.SetWill(cfg.MQTT.TopicPrefix+"/status", "offline", 1, false)
	opts.SetUsername(cfg.MQTT.Username)
	opts.SetPassword(cfg.MQTT.Password)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false)

	// TLS configuration
	if cfg.MQTT.TLSEnabled {
		opts.SetTLSConfig(createTLSConfig(cfg.MQTT.CertPath, cfg.MQTT.KeyPath))
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Publisher{
		client: client,
		config: cfg,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, msg []byte) error {
	topic := p.config.MQTT.TopicPrefix + "/normalized"
	token := p.client.Publish(topic, 1, false, msg)

	select {
	case <-token.Done():
		if token.Error() != nil {
			p.logger.Error("MQTT publish failed",
				zap.Error(token.Error()),
				zap.String("topic", topic))
			return token.Error()
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func createTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		zap.L().Error("Failed to load TLS certificate or key", zap.Error(err))
		return &tls.Config{}
	}

	caCertPool := x509.NewCertPool()
	// Optionally, load CA cert if needed:
	// caCert, err := ioutil.ReadFile("path/to/ca.crt")
	// if err == nil {
	//     caCertPool.AppendCertsFromPEM(caCert)
	// }

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
}
