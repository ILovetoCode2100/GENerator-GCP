# 🚀 Deploy to Render Right Now!

I've created everything you need to deploy. Here are your options:

## Option 1: One-Command Deploy (Recommended)

Run this single command:

```bash
./deploy-to-render-now.sh
```

This script will:

1. Check if you have Render CLI (or tell you how to install it)
2. Generate secure API keys for you
3. Ask for your Virtuoso API token
4. Deploy everything automatically

## Option 2: Manual Deploy (Even Easier!)

1. **Go to Render Dashboard**

   ```
   https://dashboard.render.com/select-repo?type=blueprint
   ```

2. **Connect Your GitHub Repo**

   - Click "Connect GitHub"
   - Select your repository
   - Render will auto-detect the `render.yaml` file

3. **Set Environment Variables**

   Add these in the Render dashboard:

   ```bash
   # Required (you need to provide these)
   VIRTUOSO_API_TOKEN = [Your Virtuoso API Token]
   API_KEYS = ["generate-with-openssl-rand-hex-32"]

   # Optional (these have defaults)
   VIRTUOSO_ORG_ID = 2242
   LOG_LEVEL = INFO
   ```

4. **Click "Create Blueprint"**

That's it! Render will build and deploy everything.

## Option 3: Deploy with GitHub

1. **Fork/Push to GitHub**

   ```bash
   git add .
   git commit -m "Add Render deployment"
   git push origin main
   ```

2. **Click Deploy Button**

   If you add this to your README, anyone can deploy with one click:

   ```markdown
   [![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/YOUR_USERNAME/virtuoso-GENerator)
   ```

## 🔑 What You Need

Only 2 things:

1. **Render Account** (free): https://render.com/register
2. **Virtuoso API Token**: Get from https://app.virtuoso.qa > Settings > API

## 📱 What Happens After Deploy?

1. **Get Your URL**

   - Format: `https://virtuoso-api-xxx.onrender.com`
   - Shows in Render dashboard after deploy

2. **Test It**

   ```bash
   # Check health
   curl https://your-app.onrender.com/health

   # View API docs
   open https://your-app.onrender.com/docs
   ```

3. **Use Your API**
   ```bash
   curl -X POST https://your-app.onrender.com/api/v1/tests/run \
     -H "X-API-Key: your-generated-key" \
     -H "Content-Type: application/json" \
     -d '{"content": "name: Test\nsteps:\n  - navigate: https://example.com\n  - assert: Example"}'
   ```

## ⏱️ Deployment Time

- First deployment: ~5-10 minutes (building Docker image)
- Subsequent deployments: ~2-3 minutes
- Auto-deploys on git push: Enabled by default

## 💰 Costs

- **Free Tier**: Perfect for testing (750 hours/month)
- **Paid**: Starts at $7/month per service
- **Estimated Total**: $0-25/month for most use cases

## 🆘 Need Help?

If you run into any issues:

1. Check Render logs: Dashboard > Your Service > Logs
2. Verify environment variables are set
3. Ensure your Virtuoso token is valid
4. Check the health endpoint: `/health?detailed=true`

## 🎯 Quick Checklist

- [ ] Have Render account
- [ ] Have Virtuoso API token
- [ ] Pushed code to GitHub
- [ ] Connected repo to Render
- [ ] Set environment variables
- [ ] Clicked deploy

That's literally all you need! The deployment is automated and production-ready. 🎉
