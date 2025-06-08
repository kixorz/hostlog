class Grid {
    name = 'Grid';
    grid = null
    path = ''
    filters = new Filters();

    getURL() {
        const url = new URL(window.location);
        url.pathname = this.path;
        const query = this.filters.getQuery();
        const params = Object.entries(query);
        params.forEach((param) => {
            const key = param[0];
            const values = param[1];
            if(Array.isArray(values)) {
                values.forEach((value) => {
                    url.searchParams.append(`${key}[]`, value);
                });
            } else {
                url.searchParams.append(key, value);
            }
        });
        return url;
    }

    load() {
        const url = this.getURL();
        fetch(url)
            .then((response) => response.text())
            .then((text) => {
                this.grid.innerHTML = text;
            })
            .catch(error => console.error(error));
    };

    init() {
        this.grid = document.getElementById('grid');
        this.path = this.grid.getAttribute('data-path');
        this.filters.init(this);
    };
};

const grid = new Grid();
document.addEventListener('DOMContentLoaded', function() {
    grid.init();
});