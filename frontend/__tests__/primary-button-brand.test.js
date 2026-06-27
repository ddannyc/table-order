/**
 * 守护：原生 <button type="primary"> 自带微信绿 #07c160，会无视 --weui-BRAND 令牌。
 * 凡是主操作按钮都必须带 weui-btn_primary 类，这样背景才走品牌粉 var(--weui-BRAND)。
 * （weui 的 .weui-btn_primary{background-color:var(--weui-BRAND)} 会盖掉原生绿。）
 */
const fs = require('fs');
const path = require('path');

const PAGES_DIR = path.join(__dirname, '..', 'pages');

function walkWxml(dir) {
  const out = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) out.push(...walkWxml(full));
    else if (entry.name.endsWith('.wxml')) out.push(full);
  }
  return out;
}

// 抓取每个 <button ...> 开标签
function buttonTags(content) {
  return content.match(/<button\b[^>]*>/g) || [];
}

describe('主操作按钮走品牌粉，不留原生微信绿', () => {
  const files = walkWxml(PAGES_DIR);

  it('存在带 type="primary" 的按钮（基线，确保用例有效）', () => {
    const total = files.reduce(
      (n, f) => n + buttonTags(fs.readFileSync(f, 'utf8')).filter((t) => /type="primary"/.test(t)).length,
      0,
    );
    expect(total).toBeGreaterThan(0);
  });

  for (const file of files) {
    const rel = path.relative(PAGES_DIR, file);
    const primaryBtns = buttonTags(fs.readFileSync(file, 'utf8')).filter((t) => /type="primary"/.test(t));
    primaryBtns.forEach((tag, i) => {
      it(`${rel} 第${i + 1}个 primary 按钮带 weui-btn_primary: ${tag.slice(0, 70)}`, () => {
        expect(tag).toMatch(/class="[^"]*\bweui-btn_primary\b[^"]*"/);
      });
    });
  }
});
