(function () {

  $('.match-height').matchHeight();

  $('#nav-mobile-toggle').click(function () {
    $('#header-nav').toggleClass('on');
  });

  $('ul#header-nav li.' + $('body').attr('id')).addClass('selected');

  $('ul#submenu li.' + $('div.sub-header').attr('id')).addClass('active');

  $('nav#toc-nav ul li.' + $('div.documentation-container').attr('id')).addClass('selected');
  
  if (localStorage.getItem("darkmode") === 'night'){
    $('div.highlighter-rouge').addClass('highlighter-rouge-dark');
  }

  function addCopyButtons(clipboard) {
    document.querySelectorAll('pre > code').forEach(function (codeBlock) {
        var button = document.createElement('div');
        button.className = 'copy-code-button';
        button.type = 'button';
        button.innerHTML = '<span class="icon" aria-label="Copy"><svg width="24" height="24" viewBox="0 0 24 24" focusable="false" role="presentation"><g fill="currentColor"><path d="M10 19h8V8h-8v11zM8 7.992C8 6.892 8.902 6 10.009 6h7.982C19.101 6 20 6.893 20 7.992v11.016c0 1.1-.902 1.992-2.009 1.992H10.01A2.001 2.001 0 0 1 8 19.008V7.992z"></path><path d="M5 16V4.992C5 3.892 5.902 3 7.009 3H15v13H5zm2 0h8V5H7v11z"></path></g></svg></span><span class="text" aria-label="Copy">Copy</span>';

        button.addEventListener('click', function () {
            clipboard.writeText(codeBlock.innerText).then(function () {
                /* Chrome doesn't seem to blur automatically,
                   leaving the button in a focused state. */
                button.blur();

                button.innerHTML = '<span class="icon" style="background-size: 8px 10px; padding-right:2px;" aria-label="Copied"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g fill="currentColor"><path d="M12 2c5.514 0 10 4.486 10 10s-4.486 10-10 10-10-4.486-10-10 4.486-10 10-10zm0-2c-6.627 0-12 5.373-12 12s5.373 12 12 12 12-5.373 12-12-5.373-12-12-12zm6.25 8.891l-1.421-1.409-6.105 6.218-3.078-2.937-1.396 1.436 4.5 4.319 7.5-7.627z"/></g></svg></span><span class="text" aria-label="Copied">Copied</span>';

                setTimeout(function () {
                    button.innerHTML = '<span class="icon" aria-label="Copy"><svg width="24" height="24" viewBox="0 0 24 24" focusable="false" role="presentation"><g fill="currentColor"><path d="M10 19h8V8h-8v11zM8 7.992C8 6.892 8.902 6 10.009 6h7.982C19.101 6 20 6.893 20 7.992v11.016c0 1.1-.902 1.992-2.009 1.992H10.01A2.001 2.001 0 0 1 8 19.008V7.992z"></path><path d="M5 16V4.992C5 3.892 5.902 3 7.009 3H15v13H5zm2 0h8V5H7v11z"></path></g></svg></span><span class="text" aria-label="Copy">Copy</span>';
                }, 1000);
            }, function (error) {
                button.innerText = 'Error';
            });
        });

        var pre = codeBlock.parentNode;
        if (pre.parentNode.classList.contains('highlight')) {
            var highlight = pre.parentNode;
            highlight.parentNode.insertBefore(button, highlight);
        } else {
            pre.parentNode.insertBefore(button, pre);
        }
    });
  }

  if (navigator && navigator.clipboard) {
    addCopyButtons(navigator.clipboard);
  } else {
      var script = document.createElement('script');
      script.src = 'https://cdnjs.cloudflare.com/ajax/libs/clipboard-polyfill/2.7.0/clipboard-polyfill.promise.js';
      script.integrity = 'sha256-waClS2re9NUbXRsryKoof+F9qc1gjjIhc2eT7ZbIv94=';
      script.crossOrigin = 'anonymous';
      script.onload = function() {
          addCopyButtons(clipboard);
      };

      document.body.appendChild(script);
  }

  $('a').click(function() {
    ga('send', 'event', {
      eventCategory: 'Click',
      eventAction: this.href,
      eventLabel: this.innerText
    });
  });

  $('#tipue_search_input').tipuesearch();
  
})();