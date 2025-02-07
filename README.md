# 🚀 Smart Money Hunter 

## Overview  
This project is a **high-performance on-chain copy trading system** that enables real-time tracking and mirroring of trades from target addresses on **EVM-compatible blockchains**. It leverages **Golang** for backend development, integrates third-party services like **1inch SDK** and **OKLink API**, and supports **multi-wallet** execution.  

## Features  
✅ **Real-Time Trade Subscription** – Uses **Redis as a message queue** to subscribe to and process target addresses' trades instantly.  
✅ **Multi-Wallet Support** – Enables **diversified trading execution** by distributing trades across multiple wallets.  
✅ **Optimized Transaction Execution** – Uses **1inch SDK** to find the best swap routes and minimize slippage.  
✅ **Secure Private Key Storage** – Implements **AES encryption** or **cloud-based KMS** for private key security.  
✅ **Scalable Data Storage** – Uses **MySQL** to persist transaction history and analytics data.  
✅ **Mempool Monitoring** – Captures trade signals before they are confirmed on-chain, enabling faster execution.  
✅ **Risk Management** – Supports stop-loss, max-drawdown limits, and wallet allocation strategies.  

## Tech Stack  
- **Language**: Golang  
- **Message Queue**: Redis (for trade subscription and real-time processing)  
- **Database**: MySQL (for trade records and analytics)  
- **Blockchain Services**: 1inch SDK, OKLink API  
- **Security**: AES-256 encrypted private key storage / KMS-based signing  
- **Infrastructure**: Docker, Kubernetes (optional for scaling)  

## Architecture  
```plaintext
┌──────────────────────────────┐
│       User Interface         │
│ (CLI / Web Dashboard / API)  │
└─────────────▲────────────────┘
              │
┌─────────────▼────────────────┐
│  On-Chain Copy Trading Core   │
│  (Golang Backend)             │
├──────────────────────────────┤
│ - Redis (Message Queue)       │
│ - MySQL (Trade Storage)       │
│ - OKLink API (Trade Tracking) │
│ - 1inch SDK (Optimal Routing) │
└─────────────▲────────────────┘
              │
┌─────────────▼────────────────┐
│     EVM-Compatible Chains    │
│ (Ethereum, BSC, Polygon, etc.)│
└──────────────────────────────┘

