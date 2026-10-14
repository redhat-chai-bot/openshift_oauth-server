document.addEventListener("DOMContentLoaded", (event) => {
  const loginButton = document.getElementById('co-login-button');
  const loginForm = document.getElementById('co-login-form');

  loginForm.addEventListener('submit', () => {
    loginButton.querySelector('.pf-v6-c-spinner').classList.remove('pf-v6-u-hidden');
    loginButton.classList.add('pf-m-in-progress');
    loginButton.disabled = true;
  });

  const checkFormValidity = () => {
    const isValid = loginForm.checkValidity();
    loginButton.disabled = !isValid;
    if (isValid) {
      loginButton.classList.remove('pf-m-in-progress');
      loginButton.querySelector('.pf-v6-c-spinner').classList.add('pf-v6-u-hidden');
    }
  };

  checkFormValidity();

  loginForm.addEventListener('change', checkFormValidity);
  document.getElementById('inputUsername').addEventListener('input', checkFormValidity);
  document.getElementById('inputPassword').addEventListener('input', checkFormValidity);
  window.addEventListener('pageshow', checkFormValidity);
});
