pragma solidity ^0.8.8;

import "../../../suave-geth/suave/sol/libraries/Suave.sol";

contract StoreKey {
    struct HintOrder {
        Suave.BidId id;
        bytes hint;
    }

    event HintEvent(Suave.BidId id, bytes hint);

    function storePrivateKey() external view returns (bytes memory) {
        address[] memory allowedList = new address[](1);
        allowedList[0] = address(this);

        Suave.Bid memory bid = Suave.newBid(10, allowedList, allowedList, "namespace");

        bytes memory _privateKey = Suave.confidentialInputs();
        Suave.confidentialStore(bid.id, "privateKey", abi.encode(_privateKey));

        HintOrder memory hintOrder;
        hintOrder.id = bid.id;
        hintOrder.hint = _privateKey;
        return abi.encodeWithSelector(this.emitHint.selector, hintOrder);
    }

    function emitHint(HintOrder memory order) public payable {
        emit HintEvent(order.id, order.hint);
    }

    function emitStoredEvent(Suave.BidId bidId) public view returns (bytes memory) {
        bytes memory _privateKey = Suave.confidentialRetrieve(bidId, "privateKey");
        HintOrder memory hintOrder;
        hintOrder.id = bidId;
        hintOrder.hint = _privateKey;
        return abi.encodeWithSelector(this.emitHint.selector, hintOrder);
    }
}
