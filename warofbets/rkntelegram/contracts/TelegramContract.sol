pragma solidity ^0.4.19;

contract TelegramContract {

    address public owner = 0x0071A7250ba11E086b0de4aFC6e0264061D8c941;
    address admin = 0x0071A7250ba11E086b0de4aFC6e0264061D8c941;
    address[] users;

    uint256 weiAmount;

    // добавить время

    mapping (address => uint256) telegramSupporters;

    bool isFinish = false;

    modifier onlyOwner() {
        if (msg.sender == owner) _;
    }

    modifier onlyAdmin() {
        if (msg.sender == admin) _;
    }

    modifier alive() {
        if(isFinish == false) _;
        // проверка на время
    }

    // Конструктор смарт-контракта
    function TelegramContract(address _admin) public {
        owner = msg.sender;
        admin = _admin;
    }

    // функция платежа
    function bet(address user) public payable {
        telegramSupporters[user] = msg.value;
        users.push(user);
        weiAmount += msg.value;
    }

    // Функция перевода всех ETH на счёт админа для последующей передачи победителям
    function transferToBetService(address service) public onlyAdmin {
        service.transfer(weiAmount);
    }

    // Возврат средств пользователям, если они выиграли
    function refund() public onlyAdmin {

        for(uint i; i < users.length; i++){
            users[i].transfer(telegramSupporters[users[i]]);
        }

    }

    function () external payable {
        // проверка на время, если прошло, то на счёт авторов
        bet(msg.sender);
    }
}