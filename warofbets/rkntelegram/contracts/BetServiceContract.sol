pragma solidity ^0.4.19;

import "node_modules/zeppelin-solidity/contracts/math/SafeMath.sol";
import "./RKNContract.sol";
import "./TelegramContract.sol";

contract BetServiceContract {

    address owner = 0x0071A7250ba11E086b0de4aFC6e0264061D8c941;
    address admin = 0x0071A7250ba11E086b0de4aFC6e0264061D8c941;

    uint256 weiFee;
    uint256 weiFund;
    uint256 weiPiece;
    uint256 otherFee;

    address[] telegramSupporters;
    address[] rknSupporters;

    mapping (address => uint256) telegramUsersWithBalance;
    mapping (address => uint256) rknUsersWithBalance;

    TelegramContract tel = new TelegramContract(admin);
    RKNContract rkn = new RKNContract(admin);

    modifier onlyAdmin() {
        if (msg.sender == admin) _;
    }

    function BetServiceContract() public {
        owner = msg.sender;
        admin = msg.sender;
    }

    // проверка на время - что время условия закончилось
    function finishHim(bool result) public onlyAdmin {
        // Если true, то выиграл Telegram, если false, то РКН
        if (result) {

            // получение денег от проигравших
            rkn.transferToBetService(address(this));

            // вычет fee и отправка авторам проекта
            owner.transfer((weiFund / 100) * 10);

            // расчёт того, сколько wei отправить победителям
            weiPiece = weiFund / telegramSupporters.length;

            tel.refund();

            // отправка ETH победителям
            for (uint i = 0; i < telegramSupporters.length; i++){
                telegramSupporters[i].transfer(weiPiece);
            }

        } 
        
        if (!result) {
            // получение денег от проигравших
            tel.transferToBetService(address(this));

            // вычет fee и отправка авторам проекта
            owner.transfer((weiFund / 100) * 10);

            // расчёт того, сколько wei отправить победителям
            weiPiece = weiFund / rknSupporters.length;

            rkn.refund();

            // отправка ETH победителям
            for (i = 0; i < rknSupporters.length; i++){
                rknSupporters[i].transfer(weiPiece);
            }
        }

        owner.transfer(otherFee);
    }

    function () external payable {
        // Если после проведения, то на счёт авторов проекта
        otherFee = msg.value;
    }
}