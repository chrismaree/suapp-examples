import { ethers, JsonRpcProvider } from "ethers";
import * as fs from "fs";

import artifact from "../../out/is-confidential.sol/IsConfidential.json";

const main = async () => {
  console.log("RUNNING SCRIPT");

  // Define the Private Key and RPC URL
  const privateKey =
    "91ab9a7e53c220e6210460b65a7a3bb2ca181412a8a7b43ff336b3df1737ce12";
  const rpcUrl = "http://localhost:8545";

  // Create a Wallet Instance
  const wallet = new ethers.Wallet(privateKey);

  // Connect Wallet to RPC
  const provider = new JsonRpcProvider(rpcUrl);
  const signer = wallet.connect(provider);

  // Create Contract Instance

  // Deploy a New Contract Instance
  console.log("Running script. Block number:", await provider.getBlockNumber());

  const contractFactory = new ethers.ContractFactory(
    artifact.abi,
    artifact.bytecode,
    signer
  );
  console.log("Deploying contract...");

  const contract = await contractFactory.deploy();
  await contract.waitForDeployment();

  console.log("Contract deployed to:", await contract.getAddress());

  console.log("Running script. Block number:", await provider.getBlockNumber());

  console.log("Calling example()...");
  const tx = await (contract as any).exampleNotConfidential();
  console.log("tx", tx);

  
};

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
