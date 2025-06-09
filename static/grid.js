class Grid {
    name = 'Grid';
    grid = null
    path = ''
    filters = new Filters();
    pagination = new Pagination();

    getURL() {
        const url = new URL(window.location);
        url.pathname = this.path;

        this.filters.updateURL(url);
        this.pagination.updateURL(url);

        return url;
    }

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
    };
};

const grid = new Grid();
document.addEventListener('DOMContentLoaded', function() {
    grid.init();
});
