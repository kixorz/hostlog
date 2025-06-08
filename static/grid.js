class Grid {
    name = 'Grid';
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
        fetch(url).then((response) => {
            console.info(response);
        });
    };

    init() {
        const grid = document.getElementById('grid');
        this.path = grid.getAttribute('data-path');
        this.filters.init(this);
    };
};

const grid = new Grid();
document.addEventListener('DOMContentLoaded', function() {
    grid.init();
});