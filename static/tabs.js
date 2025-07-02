document.addEventListener('DOMContentLoaded', () => {
    const tabLinks = document.querySelectorAll('.tab-link');
    const tabContents = document.querySelectorAll('.tab-content');

    tabLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();

            tabContents.forEach(content => {
                content.style.display = 'none';
            });

            const tabId = e.target.getAttribute('data-tab');
            document.getElementById(tabId).style.display = 'block';
        });
    });

    // dropdown

});