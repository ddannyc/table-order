/**
 * 守护：价格/金额用 Baloo 2 圆体数字（设计稿）。
 * Baloo 2 数字子集以 base64 ttf 内嵌在 app.wxss 的 @font-face，
 * 经 --font-number 令牌应用到各页价格/金额元素；中文仍走系统黑体。
 */
const fs = require('fs');
const path = require('path');

const read = (rel) => fs.readFileSync(path.join(__dirname, '..', rel), 'utf8');
const app = read('app.wxss');

describe('Baloo 2 数字字体基建（app.wxss）', () => {
  it('内嵌 @font-face：Baloo2Num + base64 truetype（无需网络，免域名白名单）', () => {
    const face = app.match(/@font-face\s*\{[^}]*\}/s);
    expect(face).not.toBeNull();
    expect(face[0]).toMatch(/font-family:\s*'Baloo2Num'/);
    expect(face[0]).toMatch(/src:\s*url\('data:font\/ttf;base64,[A-Za-z0-9+/=]{500,}'\)\s*format\('truetype'\)/);
  });

  it('定义 --font-number 令牌，回退系统字体', () => {
    expect(app).toMatch(/--font-number:\s*'Baloo2Num',[^;]*sans-serif/);
  });
});

describe('各页价格/金额套用 --font-number', () => {
  const cases = [
    ['pages/home/index.wxss', '.promo-pr'],
    ['pages/menu/index.wxss', '.menu-price'],
    ['pages/menu/index.wxss', '.menu-cart-total'],
    ['pages/order-confirm/index.wxss', '.total-amount'],
    ['pages/order-confirm/index.wxss', '.submit-amount'],
    ['pages/order-confirm/index.wxss', '.oc-item-price'],
    ['pages/profile/index.wxss', '.balance-amount'],
    ['pages/profile/index.wxss', '.stat-value'],
  ];
  it.each(cases)('%s %s 用 var(--font-number)', (file, sel) => {
    const css = read(file);
    const rule = css.match(new RegExp('\\' + sel + '\\s*\\{([^}]*)\\}'));
    expect(rule).not.toBeNull();
    expect(rule[1]).toMatch(/font-family:\s*var\(--font-number\)/);
  });
});
