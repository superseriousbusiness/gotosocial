# Custom CSS (Advanced)

CSS (Cascading Style Sheets) is a coding language used alongside HTML, which determines how a web page looks in a web browser:

> While HTML is used to define the structure and semantics of your content, CSS is used to style it and lay it out. For example, you can use CSS to alter the font, color, size, and spacing of your content, split it into multiple columns, or add animations and other decorative features.
>
> -- [Learn CSS (Mozilla)](https://developer.mozilla.org/en-US/docs/Learn/CSS)

Depending on the settings configured by the admin of your GoToSocial instance, you may be able to upload custom CSS for your account via the User Settings Panel.

This allows you to customize the appearance of your GoToSocial profile for users visiting it using a web browser.

## Example - Changing Background Color

Here's a standard GoToSocial profile page:

![A GoToSocial test profile page. The standard color scheme of grey, blue, and orange.](./../public/cssstandard.png)

Let's say we want the background color to be black instead of grey.

In the User Settings Panel, we enter the following CSS code in the Custom CSS field:

```css
.page {
  background: black;
}
```

We then click on Save Profile Info.

If we go back to our profile page and refresh the page, it now looks like this:

![The same GoToSocial test profile page. The background is now black.](./../public/cssblack.png)

If we want to get really fancy, we can add an ombre effect to the background, by using the following CSS code instead:

```css
.page {
  background: linear-gradient(crimson, purple);
}
```

After saving the css and refreshing the profile page, the profile now looks like this:

![The same GoToSocial test profile page. The background now starts dark red and fades to purple further down the page.](./../public/cssgradient.png)

## Accessibility

The importance of accessible HTML and CSS cannot be overstated. From W3:

> The Web is fundamentally designed to work for all people, whatever their hardware, software, language, location, or ability. When the Web meets this goal, it is accessible to people with a diverse range of hearing, movement, sight, and cognitive ability.
>
> Thus the impact of disability is radically changed on the Web because the Web removes barriers to communication and interaction that many people face in the physical world. However, when websites, applications, technologies, or tools are badly designed, they can create barriers that exclude people from using the Web.
>
> Accessibility is essential for developers and organizations that want to create high-quality websites and web tools, and not exclude people from using their products and services.
>
> -- [Introduction To Web Accessibility](https://www.w3.org/WAI/fundamentals/accessibility-intro/)

The standard GoToSocial theme is designed with web accessibility in mind, especially when it comes to layout, color contrasts, font sizes, and so on.

If you write custom CSS for your profile, it is very important that you make sure that it remains legible and that it behaves as expected. Buttons should look like buttons, links should look like links, text should be presented in a readable font, elements should not jump around the page, etc. Web pages can be pretty and exciting without sacrificing readability, or making things overcomplicated.

If you change your color scheme, it's a good idea to validate the new colors to make sure that they have sufficient contrast to be readable by people with visual impairments like color blindness. Once you've updated your CSS, try entering your profile URL in a contrast checking tool, like the [Color Contrast Accessibility Validator](https://color.a11y.com/Contrast). You can also use the 'Accessibility' tab in the developer tools of your web browser to check for any issues.

Styling with accessibility in mind makes the web better for everyone! Have a look at the links below for more information.

## Useful Links

- [Learn CSS (Mozilla)](https://developer.mozilla.org/en-US/docs/Learn/CSS)
- [CSS Tutorial (W3 Schools)](https://www.w3schools.com/Css/default.asp)
- [CSS and JavaScript Accessibility Best Practices (Mozilla)](https://developer.mozilla.org/en-US/docs/Learn/Accessibility/CSS_and_JavaScript#css)
- [WAVE Web Accessibility Evaluation Tool](https://wave.webaim.org/)
- [Color Contrast Accessibility Validator](https://color.a11y.com/Contrast)
