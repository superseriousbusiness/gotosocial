# 自定义 CSS（进阶）

CSS（级联样式表）是一种与 HTML 一起使用的编码语言，它决定了网页在浏览器中的外观：

> HTML 用于定义内容的结构和语义，CSS 用于对其进行样式化和布局。例如，你可以使用 CSS 来更改内容的字体、颜色、大小和间距，分成多列，或添加动画和其他装饰功能。
>
> -- [学习 CSS (Mozilla)](https://developer.mozilla.org/zh-CN/docs/Learn/CSS)

根据你的 GoToSocial 实例管理员配置的设置，你可以通过用户设置面板上传自定义 CSS 到你的账户。

这允许你为使用网络浏览器访问你的 GoToSocial 账户页的用户自定义页面外观。

## 示例 - 更改背景颜色

这是一个标准的 GoToSocial 账户页面：

![一个 GoToSocial 测试账户页面。标准配色方案是灰色、蓝色和橙色。](./../public/cssstandard.png)

假设我们想将背景颜色改为黑色而不是灰色。

在用户设置面板中，我们在自定义 CSS 字段中输入以下 CSS 代码：

```css
.page {
  background: black;
}
```

然后我们点击保存账户信息。

如果我们返回到账户页面并刷新页面，现在它看起来是这样的：

![同一个 GoToSocial 测试账户页面。背景现在是黑色。](./../public/cssblack.png)

如果我们想要更炫一点，可以使用下面的 CSS 代码为背景添加渐变效果：

```css
.page {
  background: linear-gradient(crimson, purple);
}
```

保存 CSS 并刷新账户页面后，页面现在看起来是这样的：

![同一个 GoToSocial 测试账户页面。背景现在从深红色开始，向下渐变为紫色。](./../public/cssgradient.png)

## 可访问性

可访问的 HTML 和 CSS 的重要性不容忽视。以下节选自 W3：

> 网络的基本设计理念是为所有人服务，无论他们的硬件、软件、语言、位置或能力如何。当网络达到这一目标时，它对具有不同的听力、行动、视力和认知能力的人来说都是可访问的。
>
> 因此，残疾带来的影响在网络上有根本性的不同，因为网络消除了许多人在物理世界中面临的交流和互动障碍。然而，当网站、应用程序、技术或工具设计不佳时，它们可能会带来把人排除在网络之外的新障碍。
>
> 对于希望创建高质量网站和网络工具，而不希望把使用其产品和服务的部分人群排除在外的开发者和组织来说，可访问性是必不可少的。
>
> -- [网络可访问性介绍](https://www.w3.org/WAI/fundamentals/accessibility-intro/)

标准的 GoToSocial 主题在设计中考虑了网络可访问性，特别是在布局、颜色对比、字体大小等方面。

如果你为账户编写自定义 CSS，非常重要的一点是确保它保持可读，并按预期运行。按钮应看起来像按钮，链接应看起来像链接，文本应以可读的字体呈现，元素不应在页面上跳动等。网页可以做到漂亮且令人兴奋，而不必牺牲可读性或让事情过于复杂。

如果你更改你的配色方案，最好验证新颜色，以确保它们具有足够的对比度以便视觉障碍（如色盲）的人可以阅读。一旦更新了 CSS，试着将你的账户 URL 输入对比度检查工具中，如[颜色对比度可访问性验证器](https://color.a11y.com/Contrast)。你还可以使用网络浏览器的“可访问性”选项卡检查是否存在任何问题。

以可访问性为中心进行样式设置可以让网络对所有人更友好！查看下面的链接以获取更多信息。

## 有用链接

- [学习 CSS (Mozilla)](https://developer.mozilla.org/zh-CN/docs/Learn/CSS)
- [CSS 教程 (W3 Schools)](https://www.w3schools.com/Css/default.asp)
- [CSS 和 JavaScript 可访问性最佳实践 (Mozilla)](https://developer.mozilla.org/en-US/docs/Learn/Accessibility/CSS_and_JavaScript#css)
- [WAVE 网页可访问性评估工具](https://wave.webaim.org/)
- [颜色对比度可访问性验证器](https://color.a11y.com/Contrast)
