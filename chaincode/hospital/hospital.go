package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Asset defines the structure of the asset
type Asset struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Status string `json:"status"` // e.g., "available", "checked out"
}

// SmartContract provides functions for managing hospital assets
type SmartContract struct {
	contractapi.Contract
}

// InitLedger initializes the ledger with some assets
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Owner: "owner1", Status: "available"},
		{ID: "asset2", Owner: "owner2", Status: "checked out"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return fmt.Errorf("failed to marshal asset: %v", err)
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put asset %s: %v", asset.ID, err)
		}
	}

	return nil
}

// CreateAsset adds a new asset to the ledger
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, owner string, status string) error {
	asset := Asset{
		ID:     id,
		Owner:  owner,
		Status: status,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset: %v", err)
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// QueryAsset retrieves an asset by ID
func (s *SmartContract) QueryAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal asset: %v", err)
	}

	return &asset, nil
}

// GetQueryResult retrieves all assets from the ledger
func (s *SmartContract) GetQueryResult(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	queryString := `{"selector":{}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate results: %v", err)
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// Delete removes an asset from the ledger
func (s *SmartContract) Delete(ctx contractapi.TransactionContextInterface, id string) error {
	return ctx.GetStub().DelState(id)
}

// UpdateOwner updates the owner of an asset
func (s *SmartContract) UpdateOwner(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.QueryAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset: %v", err)
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// TransferAsset transfers an asset from one owner to another
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	return s.UpdateOwner(ctx, id, newOwner)
}

// AssetExists checks if an asset exists
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to check asset existence: %v", err)
	}
	return assetJSON != nil, nil
}

// GetCallerIdentity retrieves the identity of the caller (MSP ID not directly accessible)
func (s *SmartContract) GetCallerIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	id, _ := ctx.GetStub().GetCreator()
	return fmt.Sprintf("Identity: %x", id), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
