class Filters {
    name = 'Filters';
    grid = null;
    values = new Set();

    updateURL(url) {
        const query = this.getQuery();
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
    };

    getQuery() {
        return {
            hosts: [...this.values]
        };
    };

    update() {
        this.grid.load();

        const filtersContainer = document.getElementById('active-host-filters');
        const tagsContainer = filtersContainer.querySelector('.tags');
        [...tagsContainer.children].forEach(c => c.remove());

        const activeFilters = this.values;
        if (activeFilters.size > 0) {
            filtersContainer.style.display = 'block';

            activeFilters.forEach((host) => {
                const tag = document.createElement('span');
                tag.className = 'tag is-info';
                tag.textContent = host;

                const deleteButton = document.createElement('button');
                deleteButton.setAttribute('data-host', host);
                deleteButton.className = 'delete is-small';
                deleteButton.addEventListener('click', (event) => this.toggleHandler(event));

                tag.appendChild(deleteButton);
                tagsContainer.appendChild(tag);
            });
        } else {
            filtersContainer.style.display = 'none';
        }
    };

    toggle(value) {
        if(this.values.has(value)) {
            this.values.delete(value);
        } else {
            this.values.add(value);
        }
    };

    toggleHandler(event) {
        event.preventDefault();
        const value = event.target.getAttribute('data-host');
        if(value.length) {
            this.toggle(value);
        } else {
            this.values.clear();
        }
        this.update();
    };

    register(control) {
        control.addEventListener('click', (event) => this.toggleHandler(event));
    };

    init(grid) {
        this.grid = grid;
        const controls = document.getElementsByClassName('host-filter-item');
        for(let i = 0; i < controls.length; ++i) {
            const control = controls[i];
            this.register(control);
        }
    };
};