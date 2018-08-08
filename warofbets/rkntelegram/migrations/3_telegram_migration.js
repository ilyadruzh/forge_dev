var TelegramContract = artifacts.require("./TelegramContract.sol");

module.exports = function(deployer) {
  deployer.deploy(TelegramContract);
};
