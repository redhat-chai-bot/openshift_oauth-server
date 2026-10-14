const updateTheme = () => {
  document.documentElement.classList.toggle(
    "pf-v6-theme-dark",
    window.matchMedia('(prefers-color-scheme: dark)').matches
  );
  document.documentElement.classList.toggle(
    "pf-v6-theme-high-contrast",
    window.matchMedia('(prefers-contrast: more)').matches
  );
};

updateTheme();

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme);
window.matchMedia('(prefers-contrast: more)').addEventListener('change', updateTheme);
