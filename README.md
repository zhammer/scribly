# scribly

write stories together

### tech goals
- traditional web app
- accessible/a11y-friendly website
- js not required for basic functionality
- expressive architecture/code
- tests that ensure core functionality, aren't too invasive, and improve understandibility of app

## Adding a New Theme

Scribly supports multiple visual themes that users can toggle between. To add a new theme:

1. **Add theme CSS** to `embed/static/style.css`:
   ```css
   /* Theme: My New Theme */
   body.theme-mynewtheme {
     background-color: yourcolor;
     color: yourtextcolor;
     /* ... other styles ... */
   }

   /* Override other elements as needed */
   body.theme-mynewtheme input,
   body.theme-mynewtheme textarea {
     color: yourtextcolor;
     border-color: yourbordercolor;
   }
   ```

2. **Add theme to rotation array** in `cmd/site/template.go`:
   ```go
   var availableThemes = []Theme{
       {Name: "default", CSSClass: "", Icon: "⬜"},
       {Name: "candlelit", CSSClass: "theme-candlelit", Icon: "🕯️"},
       {Name: "mynewtheme", CSSClass: "theme-mynewtheme", Icon: "🌙"}, // Your new theme
   }
   ```

   The `Icon` field is the emoji shown on the button to switch **TO** this theme (from the previous theme in rotation).

3. **Rebuild** the app since templates and static assets are embedded:
   ```bash
   go build -o scribly ./cmd/site
   ```

That's it! The theme will now appear in the rotation when users click the theme toggle button on the /me page.
