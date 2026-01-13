class Grid {
    name = 'Grid';
    grid = null;
    path = '';
    filters = new Filters();
    pagination = new Pagination();

    getURL() {
        const url = new URL(window.location);
        url.pathname = this.path;

        this.filters.updateURL(url);
        this.pagination.updateURL(url);

        return url;
    };

    load() {
        const url = this.getURL();
        fetch(url)
            .then((response) => response.text())
            .then((text) => {
                this.grid.innerHTML = text;
                this.pagination.init(this);
            })
            .catch(error => console.error(error));
    };

    init() {
        this.grid = document.getElementById('grid');
        this.path = this.grid.getAttribute('data-path');
        this.filters.init(this);
        this.pagination.init(this);
        this.initEvents();
    };

    initEvents() {
        const eventSource = new EventSource('/events');
        eventSource.onmessage = (event) => {
            if (this.pagination.page !== 0) {
                return;
            }

            const tbody = this.grid.querySelector('tbody');
            if (tbody) {
                const temp = document.createElement('template');
                temp.innerHTML = event.data;
                const row = temp.content.firstChild;

                const source = row.getAttribute('data-source');
                if (this.filters.values.size > 0 && !this.filters.values.has(source)) {
                    return;
                }

                // If there was a "No logs found" message, remove it
                if (tbody.children.length === 1 && tbody.querySelector('td[colspan="4"]')) {
                    tbody.innerHTML = '';
                }

                tbody.insertBefore(row, tbody.firstChild);

                // Optional: limit to 100 rows
                if (tbody.children.length > 100) {
                    tbody.removeChild(tbody.lastChild);
                }
            }
        };

        eventSource.onerror = (error) => {
            console.error('Server sent event error:', error);
            eventSource.close();
            setTimeout(() => this.initEvents(), 5000);
        };
    };
};

const grid = new Grid();
document.addEventListener('DOMContentLoaded', () => {
    grid.init();
});
