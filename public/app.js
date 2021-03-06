
$("[type='number']").keypress(function (evt) {
    evt.preventDefault();
});

const mrmContractABI =  [ { "anonymous": false, "inputs": [ { "indexed": false, "name": "rider", "type": "address" }, { "indexed": false, "name": "car", "type": "address" } ], "name": "RideFinished", "type": "event" }, { "constant": false, "inputs": [ { "name": "chosenRider", "type": "address" } ], "name": "acceptRideRequest", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": false, "inputs": [], "name": "cancelRideRequest", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": false, "inputs": [], "name": "finishRide", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": false, "inputs": [ { "name": "from", "type": "string" }, { "name": "to", "type": "string" }, { "name": "amount", "type": "uint256" } ], "name": "newRideRequest", "outputs": [ { "name": "", "type": "bool" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "anonymous": false, "inputs": [ { "indexed": false, "name": "rider", "type": "address" }, { "indexed": false, "name": "from", "type": "string" }, { "indexed": false, "name": "to", "type": "string" }, { "indexed": false, "name": "amount", "type": "uint256" } ], "name": "NewRideRequest", "type": "event" }, { "anonymous": false, "inputs": [ { "indexed": false, "name": "rider", "type": "address" }, { "indexed": false, "name": "car", "type": "address" } ], "name": "RideAccepted", "type": "event" }, { "inputs": [ { "name": "moovCoinAddress", "type": "address" } ], "payable": false, "stateMutability": "nonpayable", "type": "constructor" }, { "constant": true, "inputs": [], "name": "getAvailableRides", "outputs": [ { "name": "", "type": "address[]" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "moovCoin", "outputs": [ { "name": "", "type": "address" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [ { "name": "", "type": "address" } ], "name": "rides", "outputs": [ { "name": "from", "type": "string" }, { "name": "to", "type": "string" }, { "name": "amount", "type": "uint256" }, { "name": "rideStatus", "type": "uint8" }, { "name": "carAddress", "type": "address" } ], "payable": false, "stateMutability": "view", "type": "function" } ];
const moovCoinABI = [{"constant":false,"inputs":[],"name":"corruptExchange","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"INITIAL_SUPPLY","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_subtractedValue","type":"uint256"}],"name":"decreaseApproval","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_addedValue","type":"uint256"}],"name":"increaseApproval","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}];
var mrmAddress;

ws = new WebSocket('ws://' + window.location.host + '/ws');
ws.addEventListener('message', saveAddress);
testing = false;
function saveAddress(e) {
  var msg = JSON.parse(e.data);
  if (msg.testing == "true") {
    testing = true;
    document.getElementById("get-ride-button").onclick = getTestChainRide;
    document.getElementById("non-blockchain-version").style.display = "none";
  } else {
    document.getElementById("blockchain-version").style.display = "none";
    mrmAddress = msg.mrmAddress
    document.getElementById('eth-interface').style.display = "block";
    document.getElementById("getMC").onclick = getMCs;
    document.getElementById("approve-mc-button").onclick = approveMC;
    document.getElementById("get-ride-button").onclick = getRide;
    document.getElementById("finish-ride-button").onclick = finishRide;
        // Check if Web3 has been injected by the browser:
    if (typeof web3 !== 'undefined' ) {
      // You have a web3 browser! Continue below!
      startApp(web3);
    } else {
      alert("Get METAMASK! or use non blockchain version at http://test.moovlab.online");
      $(':button').prop('disabled', true);
      document.getElementById("get-ride-debug").innerHTML = "Get chrome plugin metamask to interface with the blockchain or use non blockchain version, link below"
       // Warn the user that they need to get a web3 browser
       // Or install MetaMask, maybe with a nice graphic.
    }
  }
  ws.removeEventListener('message', saveAddress)
  ws.addEventListener('message', updateCarPosition);
}

function updateCarPosition(e) {
  var msg = JSON.parse(e.data);
  if (msg.type == "Car") {
    document.getElementById('Car'+msg.id).style.top = parseInt(msg.y)+"px"
    document.getElementById('Car'+msg.id).style.left = parseInt(msg.x)+"px"
    document.getElementById('Car'+msg.id).style.transform  = "rotate("+(parseInt(msg.orientation)+180)+"deg)";
  } else if (msg.type == "Stoplight"){
      var lightMap = {
          "0": "#c70101",
          "1": "Orange",
          "2": "Green"
      };
      document.getElementById('StopLight'+msg.id).querySelector('div[name="West"]').style.background = lightMap[msg.north];
      document.getElementById('StopLight'+msg.id).querySelector('div[name="South"]').style.background = lightMap[msg.west];
      document.getElementById('StopLight'+msg.id).querySelector('div[name="East"]').style.background = lightMap[msg.south];
      document.getElementById('StopLight'+msg.id).querySelector('div[name="North"]').style.background = lightMap[msg.east];
  } else if (msg.type == "RideStatus" && (testing || msg.address.toLowerCase() == coinbase)) {
    switch(msg.state) {
        case "To Pick Up":
            var carName = document.getElementById('Car' + msg.id).name;
            document.getElementById("get-ride-debug").innerHTML = carName + " is on the way";
            break;
        case "At Pick Up":
            var carName = document.getElementById('Car' + msg.id).name;
            document.getElementById("get-ride-debug").innerHTML = carName + " is at Pickup";
            break;
        case "At Drop Off":
            var carName = document.getElementById('Car' + msg.id).name;
            document.getElementById("get-ride-debug").innerHTML = carName + " is at Dropoff";
            if (!testing && msg.address.toLowerCase() == coinbase) {
                document.getElementById("finish-ride-button").style.visibility = "visible";
            }
            break;
    }
  }
};

window.addEventListener('load', function() {
  document.getElementById("set-start-point-button").onclick =
    function (e) {
      document.getElementById('Map').onclick = setStartPoint;
      document.getElementById("get-ride-debug").innerHTML = " Click anywhere on map to select start point";
    };
  document.getElementById("set-end-point-button").onclick =
    function (e) {
      document.getElementById('Map').onclick = setEndPoint;
      document.getElementById("get-ride-debug").innerHTML = " Click anywhere on map to select end point";
    };
})

async function startApp(web3) {
    eth = new Eth(web3.currentProvider);
    var version = await eth.net_version();
    if(version == "3") {
      mrm = eth.contract(mrmContractABI).at(mrmAddress);
      const moovCoinAddress = await mrm.moovCoin();
      moovCoin = eth.contract(moovCoinABI).at(moovCoinAddress[0]);
      coinbase = await eth.coinbase();
      updateView();
    } else {
      document.getElementById("ui").innerHTML = "Change network to Ropsten";
    }
}

async function updateView() {
    document.getElementById("mrm-address").innerHTML = mrmAddress;
    document.getElementById("user-address").innerHTML = coinbase;

    const userBalance = await moovCoin.balanceOf(coinbase);
    document.getElementById("user-balance").innerHTML = userBalance[0].toNumber();
    if (userBalance[0]==0) {
      document.getElementById("get-mc-debug").innerHTML = "| Purchase Move Coins to get a ride";
    } else {
      document.getElementById("get-mc-debug").innerHTML = "";
    }

    const userApprovedMCs = await(moovCoin.allowance(coinbase, mrmAddress))
    document.getElementById("user-approved-mcs").innerHTML = userApprovedMCs[0].toNumber();
    document.getElementById("approve-mc-field").setAttribute("max", userBalance[0].toNumber() - userApprovedMCs[0].toNumber());
    if (userBalance[0] > 0 && userApprovedMCs[0] == 0) {
      document.getElementById("approve-mc-debug").innerHTML = "  Authorize smart contract to spend money on your behalf";
    } else {
      document.getElementById("approve-mc-debug").innerHTML = "";
    }
    $('#approve-mc-button').prop('disabled', userBalance[0]==0);

    document.getElementById("get-ride-amount-field").setAttribute("max", userApprovedMCs[0].toNumber());
    $('#get-ride-button').prop('disabled', userApprovedMCs[0]==0);
    $('#set-start-point-button').prop('disabled', userApprovedMCs[0]==0);
    $('#set-end-point-button').prop('disabled', userApprovedMCs[0]==0);

    const rideState = await mrm.rides(coinbase);
    if (rideState["rideStatus"].toNumber() == 2) {
      document.getElementById("finish-ride-button").style.visibility = "visible";
    } else {
      document.getElementById("finish-ride-button").style.visibility = "hidden";
    }
}

async function waitForTxToBeMined (txHash) {
  $(':button').prop('disabled', true);
  let txReceipt
  while (!txReceipt) {
    try {
      txReceipt = await eth.getTransactionReceipt(txHash)
    } catch (err) {
      console.log("Tx Mine failed");
    }
  }
  console.log("Tx Mined");
  $(':button').prop('disabled', false);
  updateView();
}

function getMCs(){
    console.log("trying to get MC");
    const amount = parseInt(document.getElementById("get-mc-field").value);
    if(amount <= 0) {
      console.log("Come on dude, ask for more than 0");
      return;
    }
    document.getElementById("get-mc-field").value = 0;

    moovCoin.corruptExchange({from:coinbase, value:10000000000000000*amount}).then(function (txHash) {
      console.log('Transaction sent');
      console.dir(txHash);
      waitForTxToBeMined(txHash);
    });;
}

async function approveMC(){
    console.log("trying to approve");
    const amount = parseInt(document.getElementById("approve-mc-field").value);
    if(amount <= 0) {
      console.log("Come on dude, ask for more than 0");
      return;
    }
    document.getElementById("approve-mc-field").value = 0;

    moovCoin.increaseApproval(mrmAddress, amount, { from: coinbase }).then(function (txHash) {
      console.log('Transaction sent');
      console.dir(txHash);
      waitForTxToBeMined(txHash);
    });
}

async function getRide(){
    console.log("trying to get ride");
    var debugElement = document.getElementById("get-ride-debug");
    debugElement.innerHTML = "";
    const start = getLocations('start-point')
    const end = getLocations('end-point')
    const amount = parseInt(document.getElementById("get-ride-amount-field").value);

    const rideState = await mrm.rides(coinbase);
    if (rideState["rideStatus"].toNumber() != 0) {
      debugElement.innerHTML = "End current ride to get another ride";
      return;
    }

    if (start == "location not set"){
      debugElement.innerHTML = "Set Start Point before getting ride";
      return;
    }else if (end == "location not set"){
      debugElement.innerHTML = "Set End Point before getting ride";
      return;
    }else if (amount <= 0 ) {
      document.getElementById("get-ride-debug").innerHTML = "Enter a ride amount greater than 0";
      return;
    }
    console.log(start+" "+end);
    mrm.newRideRequest(start, end, amount, { from: coinbase }).then(function (txHash) {
      console.log('Transaction sent');
      console.dir(txHash);
      waitForTxToBeMined(txHash);
      document.getElementById("get-ride-amount-field").value = 0;
    });
}

async function finishRide(){
  mrm.finishRide({ from: coinbase }).then(function (txHash) {
      console.log('Transaction sent');
      console.dir(txHash);
      waitForTxToBeMined(txHash);
    });
  document.getElementById("finish-ride-button").style.visibility = "hidden";
}

function getTestChainRide() {
  console.log("trying to get ride");
    var debugElement = document.getElementById("get-ride-debug");
    debugElement.innerHTML = "";
    const start = getLocations('start-point')
    const end = getLocations('end-point')
    const amount = parseInt(document.getElementById("get-ride-amount-field").value);
    if (start == "location not set"){
      debugElement.innerHTML = "Set Start Point before getting ride";
      return;
    }else if (end == "location not set"){
      debugElement.innerHTML = "Set End Point before getting ride";
      return;
    }else if (amount <= 0 ) {
      document.getElementById("get-ride-debug").innerHTML = "Enter a ride amount greater than 0";
      return;
    }
    console.log(start+" "+end);
    ws.send(JSON.stringify({
                        from: start,
                        to: end}));
}

function getLocations(locString) {
  if (document.getElementById(locString).style.visibility == "visible") {
    var y = parseInt(document.getElementById(locString).style.top) + 10;
    y -= document.getElementById('Map').offsetTop;
    var x = parseInt(document.getElementById(locString).style.left) + 5;
    x -= document.getElementById('Map').offsetLeft;
    return ""+x+","+y;
  } else {
    return "location not set";
  }
}


function setStartPoint(e){
  document.getElementById('start-point').style.visibility = "visible";
  document.getElementById('start-point').style.top = (e.pageY-10)+"px";
  document.getElementById('start-point').style.left = (e.pageX-5)+"px";
  document.getElementById("get-ride-debug").innerHTML = "";
  document.getElementById('Map').onclick = null;
}


function setEndPoint(e){
  document.getElementById('end-point').style.visibility = "visible";
  document.getElementById('end-point').style.top = (e.pageY-10)+"px";
  document.getElementById('end-point').style.left = (e.pageX-5)+"px";
  document.getElementById("get-ride-debug").innerHTML = "";
  document.getElementById('Map').onclick = null;
}


// console.log((e.pageX-document.getElementById('Map').offsetLeft)+","+(e.pageY-document.getElementById('Map').offsetTop));