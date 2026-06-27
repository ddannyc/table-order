/**
 * 守护：主操作按钮必须走品牌粉，不能露出微信原生绿 #07c160。
 *
 * 关键坑：原生 <button type="primary"> 的绿底由微信基础组件渲染，优先级高于
 * 作者样式类——即使加了 .weui-btn_primary{background:var(--weui-BRAND)} 也盖不掉。
 * （菜单里的“选规格”是 <view class="weui-btn_primary">，view 没有原生绿才显粉。）
 * 因此正确做法是去掉 type="primary"，只留 weui-btn_primary 类，让 weui 的粉底生效；
 * type 只影响原生配色，不影响 bindtap/loading/open-type 行为。
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

function buttonTags(content) {
  return content.match(/<button\b[^>]*>/g) || [];
}

describe('主操作按钮走品牌粉，不露微信原生绿', () => {
  const files = walkWxml(PAGES_DIR);

  it('任何页面都不再用 <button type="primary">（原生绿盖不掉）', () => {
    const offenders = [];
    for (const file of files) {
      buttonTags(fs.readFileSync(file, 'utf8'))
        .filter((t) => /type="primary"/.test(t))
        .forEach((t) => offenders.push(`${path.relative(PAGES_DIR, file)}: ${t}`));
    }
    expect(offenders).toEqual([]);
  });

  it('主操作按钮改用 weui-btn_primary 类（基线：至少有几个）', () => {
    const total = files.reduce(
      (n, f) => n + buttonTags(fs.readFileSync(f, 'utf8')).filter((t) => /\bweui-btn_primary\b/.test(t)).length,
      0,
    );
    expect(total).toBeGreaterThan(0);
  });
});
