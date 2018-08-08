Web3 = require('web3')
const Eth = require('ethjs-query')
const EthContract = require('ethjs-contract')

// window.addEventListener('load', function() {
//     if (typeof web3 !== 'undefined') {
//       console.log("yes");
//       startApp(web3);
//     } else {
//       console.log("no");
//       // Warn the user that they need to get a web3 browser
//       // Or install MetaMask, maybe with a nice graphic.
//     }
// })

// function startApp(web3) {
//     const eth = new Eth(web3.currentProvider)
//     const contract = new EthContract(eth)
//     initContract(contract)
// }

// const addressTelegram = '0x0071a7250ba11e086b0de4afc6e0264061d8c941';

// var abi = [ { "constant": false, "inputs": [ { "name": "newString", "type": "string" } ], "name": "setString", "outputs": [], "payable": false, "type": "function" }, { "constant": true, "inputs": [], "name": "getString", "outputs": [ { "name": "", "type": "string", "value": "Hello World!" } ], "payable": false, "type": "function" } ];

// window.addEventListener('load', function() {
//     if (typeof web3 !== 'undefined') {
//       web3 = new Web3(web3.currentProvider);
//       console.log("yes");
//     } else {
//       web3 = new Web3(new Web3.providers.HttpProvider("http://localhost:8545")); 
//       console.log("no");
//     }
// })
  
// const addressTelegram = '0x0071a7250ba11e086b0de4afc6e0264061d8c941';
// const addressRKN = '0x0071a7250ba11e086b0de4afc6e0264061d8c941'

// const abiSrc = [{
//     "constant": false,
//     "inputs": [
//       {
//         "name": "_to",
//         "type": "address"
//       },
//       {
//         "name": "_value",
//         "type": "uint256"
//       }
//     ],
//     "name": "transfer",
//     "outputs": [
//       {
//         "name": "success",
//         "type": "bool"
//       }
//     ],
//     "payable": false,
//     "type": "function"
//   }];

// var contract = web3.eth.contract(abi);
// var stringHolder = contract.at(addressTelegram)

  // var transferFundsToTelegram = document.querySelector('.transferFundsToTelegram')
  // var transferFundsToRKN = document.querySelector('.transferFundsToTelegram')

  transferFundsToTelegram.addEventListener('click', function() {
    if (typeof web3 === 'undefined') {
      return renderMessage('<div>You need to install <a href=“https://metmask.io“>MetaMask </a> to use this feature.  <a href=“https://metmask.io“>https://metamask.io</a></div>')
    }
    var user_address = web3.eth.accounts[0]
    
    web3.eth.sendTransaction({
      to: addressTelegram,
      from: user_address,
      value: web3.toWei('1', 'ether'),
    }, function (err, transactionHash) {
      if (err) return renderMessage('There was a problem!: ' + err.message)
      renderMessage('Thanks for the generosity!!')
      })
  })

  // transferFundsToRKN.addEventListener('click', function() {
  //   if (typeof web3 === 'undefined') {
  //     return renderMessage('<div>You need to install <a href=“https://metmask.io“>MetaMask </a> to use this feature.  <a href=“https://metmask.io“>https://metamask.io</a></div>')
  //   }
  //   var user_address = web3.eth.accounts[0]
    
  //   web3.eth.sendTransaction({
  //     to: addressRKN,
  //     from: user_address,
  //     value: web3.toWei('1', 'ether'),
  //   }, function (err, transactionHash) {
  //     if (err) return renderMessage('There was a problem!: ' + err.message)
  //     renderMessage('Thanks for the generosity!!')
  //     })
  // })
  
  // function startApp(web3) {
  //   const eth = new Eth(web3.currentProvider)
  //   const contract = new EthContract(eth)
  //   toTelegram(contract)
  //   toRKN(contract)
  // }

  
//   function toTelegram (contract) {
//     const TelegramSupporters = contract(abi)
//     const telegramSupport = TelegramSupporters.at(addressTelegram)
//     sendETH(telegramSupport)
//   }

//   function toRKN (contract) {
//     const RKNSupporters = contract(abi)
//     const RKNSupport = RKNSupporters.at(addressRKN)
//     sendETH(RKNSupport)
//   }

//   function sendETH (telegramSupport, RKNSupport) {
//     var button = document.querySelector('.transferFundsToTelegram')
//     button.addEventListener('click', function() {
//       telegramSupport.transfer(toAddress, value, { from: addr }).then(function (txHash) {
//         console.log('Transaction sent')
//         console.dir(txHash)
//         waitForTxToBeMined(txHash)
//       }).catch(console.error)
//     })

//     var button = document.querySelector('.transferFundsToRKN')
//     button.addEventListener('click', function() {
//       RKNSupport.transfer(toAddress, value, { from: addr }).then(function (txHash) {
//         console.log('Transaction sent')
//         console.dir(txHash)
//         waitForTxToBeMined(txHash)
//       }).catch(console.error)
//     })

//   async function waitForTxToBeMined (txHash) {
    
//     let txReceipt 
//     while (!txReceipt) {
//       try {
//         txReceipt = await eth.getTransactionReceipt(txHash)
//         } catch (err) {
//           return indicateFailure(err)
//           }
//     }
//     indicateSuccess()
//   }
// }


