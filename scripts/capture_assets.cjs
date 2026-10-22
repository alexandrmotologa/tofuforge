const path = require('path');
const puppeteer = require(path.join(__dirname, '..', 'web', 'node_modules', 'puppeteer'));
const fs = require('fs');
const { execSync } = require('child_process');

async function capture() {
  const imagesDir = path.join(__dirname, '..', 'docs', 'images');
  if (!fs.existsSync(imagesDir)) {
    fs.mkdirSync(imagesDir, { recursive: true });
  }

  const framesDir = path.join(__dirname, '..', 'temp_frames');
  if (fs.existsSync(framesDir)) {
    fs.rmSync(framesDir, { recursive: true, force: true });
  }
  fs.mkdirSync(framesDir, { recursive: true });

  console.log('Launching browser for authentic visual assets capture...');
  const browser = await puppeteer.launch({
    headless: 'new',
    defaultViewport: { width: 1440, height: 900, deviceScaleFactor: 2 },
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  await page.goto('http://localhost:8085', { waitUntil: 'networkidle0' });

  // Wait for React Flow nodes
  await page.waitForSelector('.react-flow__node', { timeout: 10000 });
  await new Promise(r => setTimeout(r, 1200));

  let frameIdx = 0;
  const snapFrame = async () => {
    const filename = path.join(framesDir, `frame_${String(frameIdx++).padStart(3, '0')}.png`);
    await page.screenshot({ path: filename });
  };

  // 1. Initial Overview Screenshot
  console.log('Capturing Screenshot 1: Overview...');
  await page.screenshot({ path: path.join(imagesDir, 'tofuforge-1.png') });
  for (let i = 0; i < 6; i++) await snapFrame();

  // 2. Click on the RDS database node to highlight cascade and open diff modal
  console.log('Clicking on database node to highlight cascade and open attribute diff...');
  const dbNode = await page.evaluateHandle(() => {
    const nodes = document.querySelectorAll('.react-flow__node');
    for (const n of nodes) {
      if (n.textContent && n.textContent.includes('customer_orders_postgresql')) {
        return n;
      }
    }
    return nodes[0];
  });

  if (dbNode) {
    await dbNode.click();
    await new Promise(r => setTimeout(r, 800));
  }

  console.log('Capturing Screenshot 2: Attribute Diff Modal & Downstream Cascade...');
  await page.screenshot({ path: path.join(imagesDir, 'tofuforge-2.png') });
  for (let i = 0; i < 7; i++) await snapFrame();

  // 3. Close modal
  console.log('Closing diff modal and switching to Security tab...');
  const closeBtn = await page.$('button[title="Close"], button svg.lucide-x');
  if (closeBtn) {
    await closeBtn.click();
    await new Promise(r => setTimeout(r, 400));
  } else {
    await page.keyboard.press('Escape');
    await new Promise(r => setTimeout(r, 400));
  }

  // Click Security Tab
  await page.evaluate(() => {
    const buttons = document.querySelectorAll('button');
    for (const b of buttons) {
      if (b.textContent && b.textContent.includes('Security')) {
        b.click();
        return;
      }
    }
  });
  await new Promise(r => setTimeout(r, 800));

  console.log('Capturing Screenshot 3: Security Audit Tab...');
  await page.screenshot({ path: path.join(imagesDir, 'tofuforge-3.png') });
  for (let i = 0; i < 7; i++) await snapFrame();

  // 4. Open PR Comment Modal
  console.log('Opening PR Comment Modal...');
  await page.evaluate(() => {
    const buttons = document.querySelectorAll('button');
    for (const b of buttons) {
      if (b.textContent && b.textContent.includes('PR Comment')) {
        b.click();
        return;
      }
    }
  });
  await new Promise(r => setTimeout(r, 800));

  console.log('Capturing Screenshot 4: PR Comment Markdown Modal...');
  await page.screenshot({ path: path.join(imagesDir, 'tofuforge-4.png') });
  for (let i = 0; i < 7; i++) await snapFrame();

  await browser.close();

  // 5. Generate high-quality animated demo GIF with ffmpeg
  console.log('Generating animated demo GIF with ffmpeg...');
  const gifPath = path.join(imagesDir, 'tofuforge_demo.gif');
  const ffmpegCmd = `ffmpeg -y -framerate 4 -i "${framesDir}\\frame_%03d.png" -vf "fps=6,scale=960:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer" "${gifPath}"`;
  
  try {
    execSync(ffmpegCmd, { stdio: 'inherit' });
    console.log(`Animated demo GIF generated at: ${gifPath}`);
  } catch (err) {
    console.error('Error generating GIF with ffmpeg:', err.message);
  }

  // Cleanup temp frames
  try {
    fs.rmSync(framesDir, { recursive: true, force: true });
  } catch (e) {}

  console.log('All visual assets captured successfully!');
}

capture().catch(err => {
  console.error('Capture failed:', err);
  process.exit(1);
});
