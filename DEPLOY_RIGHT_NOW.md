# 🚀 Deploy RIGHT NOW - One Command!

I've created everything for you. Here's literally all you need to do:

## The One Command:

```bash
cd gcp && ./deploy-now.sh
```

## That's it! The script will:

1. ✅ Check if you have Google Cloud CLI (tells you how to install if not)
2. ✅ Open browser for Google login (just click "Allow")
3. ✅ Ask for your Virtuoso API token (get from https://app.virtuoso.qa)
4. ✅ Deploy everything automatically
5. ✅ Give you your API URL and keys

## What You Get:

In ~15 minutes, you'll have:

- 🌐 **Live API URL**: `https://virtuoso-api-xxx.run.app`
- 🔑 **API Keys**: Auto-generated for your clients
- 📚 **API Docs**: `https://virtuoso-api-xxx.run.app/docs`
- 💰 **Cost**: $0/month when idle, ~$0.50 per million requests

## If You Don't Have Google Cloud CLI:

```bash
# Mac:
brew install google-cloud-sdk

# Windows/Linux:
curl https://sdk.cloud.google.com | bash
```

Then run the deploy command above.

## Test Your Deployment:

After deployment completes, test with:

```bash
# Check health (no auth needed)
curl https://your-api-url.run.app/health

# Test API (use the key from deployment output)
curl -H "X-API-Key: your-generated-key" https://your-api-url.run.app/api/v1/commands
```

---

**That's literally everything!** Just run:

```bash
cd gcp && ./deploy-now.sh
```

Your API will be live in 15 minutes. The script handles EVERYTHING else. 🎉
