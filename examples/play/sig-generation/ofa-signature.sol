// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.8;

import "../../../suave-geth/suave/sol/libraries/Suave.sol";
import "./RLB.sol";

contract ConfidentialKeyStore {
    string private privateKey = "0x964f7e4657e386b633a056ebbdd060d45cf64b6745e65f0ccef34f072da45e61";

    struct Transaction {
        uint8 txType;
        uint64 nonce;
        uint256 maxPriorityFeePerGas;
        uint256 maxFeePerGas;
        uint64 gasLimit;
        address from;
        address to;
        uint256 value;
        bytes data;
    }

    function callback() external payable {}

    function createTransaction() public pure returns (bytes memory) {
        Transaction memory txn;
        txn.from = 0x718648C8c531F91b528A7757dD2bE813c3940608; // Set the sender's address
        txn.to = 0x9A8f92a830A5cB89a3816e3D267CB7791c16b04D;
        txn.gasLimit = 21000;
        txn.maxFeePerGas = 50 gwei; // Adjust this value based on the current gas price
        txn.maxPriorityFeePerGas = 0; // Set the maximum priority fee per gas
        txn.nonce = 0; // Replace this with the actual nonce of the sender's address
        txn.value = 420 wei;
        txn.data = bytes("");
        txn.txType = 2; // Set the type of transaction

        // Encode and pack all fields using the functions from RLB.sol
        bytes memory encodedFrom = RLPEncoder.encodeAddress(txn.from);
        bytes memory encodedTo = RLPEncoder.encodeAddress(txn.to);
        bytes memory encodedGasLimit = RLPEncoder.encodeUint64(txn.gasLimit);
        bytes memory encodedMaxFeePerGas = RLPEncoder.encodeUint256(txn.maxFeePerGas);
        bytes memory encodedMaxPriorityFeePerGas = RLPEncoder.encodeUint256(txn.maxPriorityFeePerGas);
        bytes memory encodedNonce = RLPEncoder.encodeUint64(txn.nonce);
        bytes memory encodedValue = RLPEncoder.encodeUint256(txn.value);
        bytes memory encodedData = RLPEncoder.encodeNonSingleBytesLen(uint64(txn.data.length));
        bytes memory encodedType = RLPEncoder.encodeUint64(txn.txType);

        // Concatenate all encoded fields
        bytes memory encodedTxn = abi.encodePacked(
            encodedType,
            encodedNonce,
            encodedMaxPriorityFeePerGas,
            encodedMaxFeePerGas,
            encodedGasLimit,
            encodedFrom,
            encodedTo,
            encodedValue,
            encodedData
        );

        return encodedTxn;
    }

    function signEthTransaction() external view returns (bytes memory) {
        // Call the signEthTransaction function from Suave.sol
        bytes memory txn = createTransaction();
        bytes memory signedTxn = Suave.signEthTransaction(txn, "1", privateKey);

        // If the signedTxn is empty, revert the transaction
        if (signedTxn.length == 0) {
            revert("SIGN_ETH_TRANSACTION failed");
        }

        return signedTxn;
    }
}
