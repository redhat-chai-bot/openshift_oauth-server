function codeSnippet(element, textBlocks = []) {
  const snippetString = textBlocks.join(' ').trim();

  const snippetHTML = textBlocks
    .map((text) => {
      const span = document.createElement('span');
      span.className = 'nowrap';
      span.innerText = text;
      return span.outerHTML;
    })
    .join(' ');

  const copyButton = document.createElement('button');
  copyButton.innerText = 'Copy';

  const resetCopyButton = () => {
    copyButton.innerText = 'Copy';
    copyButton.disabled = false;
  };

  copyButton.onclick = () => {
    copyButton.disabled = true;
    navigator.clipboard.writeText(snippetString)
      .then(() => {
        copyButton.innerText = 'Copied!';
        setTimeout(resetCopyButton, 3000);
      })
      .catch(() => {
        copyButton.innerText = 'Error!';
        setTimeout(resetCopyButton, 3000);

      });
  };

  const showButton = document.createElement('button');
  showButton.innerText = 'Show';

  showButton.onclick = () => {
    element.innerHTML = snippetHTML;
    element.appendChild(copyButton);
  };

  element.innerHTML = '***';
  element.appendChild(copyButton);
  element.appendChild(showButton);
}

const tokenData = document.getElementById('token-data').dataset;
codeSnippet(document.getElementById('token'), [tokenData.accessToken]);
codeSnippet(document.getElementById('login'), [
  'oc login', '--token=' + tokenData.accessToken, '--server=' + tokenData.server,
]);
codeSnippet(document.getElementById('apiCall'), [
  'curl', '-H "Authorization: Bearer ' + tokenData.accessToken + '"',
  '"' + tokenData.server + '/apis/user.openshift.io/v1/users/~"',
]);
