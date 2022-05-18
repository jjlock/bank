function init() {
    switch (window.location.pathname) {
        case '/':
            initForm('login-form', '/login', handleLoginResponse);
            break;
        case '/login':
            initForm('login-form', '/login', handleLoginResponse);
            break;
        case '/signup':
            initForm('signup-form', '/signup', handleSignupResponse);
            break;
        case '/account':
            initForm('transaction-form', '/transaction', handleTransactionResponse, true);
            initLogout();
            break;
    }
}

function initForm(formId, resource, handleResponse, multiSubmit=false) {
    const form = document.getElementById(formId);
    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const formData = new FormData(form);

        if (multiSubmit) {
            // The name/value pair of the button that triggered the
            // submit is not contained in the FormData because the
            // FormData is being constructed manually, so we need to
            // append the name/value pair to the FormData ourselves.
            const button = event.submitter;
            formData.append(button.getAttribute('name'), button.getAttribute('value'));
        }

        try {
            const response = await fetch(resource, {
                method: 'POST',
                body: formData
            });

            form.reset();
            handleResponse(response);
        } catch (error) {
            console.log(error);
        }
    });
}

function showStatus(isSuccess, text) {
    const status = document.getElementById('status');
    status.style.color = isSuccess ? 'green' : 'red';
    status.innerText = text;
}

function handleLoginResponse(response) {
    if (!response.ok) {
        showStatus(false, 'invalid_input');
        return;
    }

    let redirect = '/account';
    const returnTo = (new URL(window.location)).searchParams.get('return_to');
    if (returnTo !== null) {
        try {
            redirect = decodeURI(returnTo);
        } catch (error) {
            console.log(error);
        }
    }

    window.location.replace(redirect);
}

function handleSignupResponse(response) {
    if (!response.ok) {
        showStatus(false, 'invalid_input');
        return;
    }

    window.location.replace('/account');
}

async function handleTransactionResponse(response) {
    if (!response.ok) {
        showStatus(false, 'invalid_input');
        return;
    }

    const balance = await response.text();
    document.getElementById('current-balance').innerText = '$' + balance

    showStatus(true, 'success');
}

function initLogout() {
    document.getElementById('logout-form').addEventListener('submit', (event) => {
        event.preventDefault();
        fetch('/logout', {method: 'POST'});
        window.location.replace('/');
    });
}

init();
