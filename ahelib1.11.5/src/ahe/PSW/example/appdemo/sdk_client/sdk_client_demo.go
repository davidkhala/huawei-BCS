package sdk_client

import (
	"ahe/PSW/lib/zkproof/util"
	"errors"
	"flag"
	"fmt"

	clichannel "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = util.GetLog("UserClient")

const (
	defaultConfig        = "./config_test.yaml"
	defaultChannel       = "mychannel"
	defaultOrgID         = "org1"
	defaultOrgPath       = "org1"
	defaultTxChainCodeID = "ahe"
	defaultTxNum         = "10"
	defaultSenderAddr    = "a"
	defaultReceiverAddr  = "b"
)

type BaseSetupImpl struct {
	client          *clichannel.Client
	peers           []fab.Peer
	orgPath         string
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
}

type TransRecord struct {
	FromAddr string
	ToAddr   string
	TXType   string
	Balance  string
	TX       string
	remark   string
}

func Init() (*BaseSetupImpl, error) {

	config := flag.String("config", defaultConfig, "address of config_test.yaml")
	channel := flag.String("channel", defaultChannel, "channel name")
	orgID := flag.String("orgID", defaultOrgID, "org")
	orgPath := flag.String("orgPath", defaultOrgPath, "org ca")
	TxChainCodeID := flag.String("TxChainCodeID", defaultTxChainCodeID, "name of chainCode")
	//senderAddr := flag.String("senderAddr", defaultSenderAddr, "addr of sender")
	//receiverAddr := flag.String("receiverAddr", defaultReceiverAddr, "addr of receiver")
	//tx := flag.String("tx", defaultTxNum, "transaction amount")
	flag.Parse()

	setup := &BaseSetupImpl{
		ConfigFile:  *config,
		ChannelID:   *channel,
		OrgID:       *orgID,
		orgPath:     *orgPath,
		ChainCodeID: *TxChainCodeID,
	}

	if err := setup.Initialize(); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return setup, nil
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize() error {
	// Create SDK setup for the integration tests
	cnfg := config.FromFile(setup.ConfigFile)

	var opts []fabsdk.Option
	opts = append(opts, fabsdk.WithCorePkg(&defcore.ProviderFactory{}), fabsdk.WithOrgid(setup.OrgID))

	sdk, err := fabsdk.New(cnfg, opts...)
	if err != nil {
		fmt.Println("Error initializing SDK:", err.Error())
		return err
	}
	ctx, err := sdk.Context()()
	if err != nil {
		fmt.Println(err)
		return err
	}
	endpointConfig := ctx.EndpointConfig()
	peersConfig, _ := endpointConfig.PeersConfig(setup.OrgID)
	var peers []fab.Peer
	for _, p := range peersConfig {
		endorser, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: p})
		if err != nil {
			fmt.Println("Failed to create peers", err)
		}
		peers = append(peers, endorser)
	}
	setup.peers = peers

	mspClient, err := msp.New(sdk.Context(), msp.WithOrg(setup.OrgID))
	if err != nil {
		fmt.Println("error creating MSP client:", err.Error())
		return err
	}

	user, err := mspClient.GetSigningIdentity("Admin")
	if err != nil {
		fmt.Println("GetSigningIdentity returned error: ", err.Error())
		return err
	}

	session := sdk.Context(fabsdk.WithIdentity(user))

	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(session, setup.ChannelID)
	}
	channelClient, err := clichannel.New(channelProvider)
	if err != nil {
		fmt.Println("Error getting channel client: ", err.Error())
		return err
	}
	setup.client = channelClient
	return nil
}

func Query(setup *BaseSetupImpl, fcn string, args [][]byte) ([]*fab.TransactionProposalResponse, error) {
	var qureyOpts []clichannel.RequestOption
	if len(setup.peers) > 0 {
		qureyOpts = append(qureyOpts, clichannel.WithTargets(setup.peers[0]))
	}

	response, err := setup.client.Query(
		clichannel.Request{
			ChaincodeID: setup.ChainCodeID,
			Fcn:         fcn,
			Args:        args,
		},
		qureyOpts...,
	)
	if err != nil {
		return nil, errors.New("Failed to qurey a value")
	}

	return response.Responses, nil
}

func Invoke(setup *BaseSetupImpl, fcn string, args [][]byte) ([]*fab.TransactionProposalResponse, error) {
	var invokeOpts []clichannel.RequestOption
	if len(setup.peers) > 0 {
		invokeOpts = append(invokeOpts, clichannel.WithTargets(setup.peers[0]))
	}

	response, err := setup.client.Execute(
		clichannel.Request{
			ChaincodeID: setup.ChainCodeID,
			Fcn:         fcn,
			Args:        args,
		},
		invokeOpts...,
	)
	if err != nil {
		return nil, errors.New("Failed to invoke value")
	}

	if response.TxValidationCode == pb.TxValidationCode_VALID {
		fmt.Println("Invoke Chaincode successfully")
		return response.Responses, nil
	} else {
		return nil, errors.New("Invoke Chaincode failed")
	}
}

func (setup *BaseSetupImpl) Peers() []fab.Peer {
	return setup.peers
}
