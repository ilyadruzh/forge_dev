var BetService = artifacts.require("./BetServiceContract.sol");

module.exports = function(deployer) {
  deployer.deploy(BetService);
};
