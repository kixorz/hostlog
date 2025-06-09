class Pagination {
    name = 'Pagination';
    page = 0;

    updateURL(url) {
        url.searchParams.append('page', this.getPage());
    };

    getPage() {
        return this.page;
    };

    init(grid) {
        this.grid = grid;
        const paginationLinks = this.grid.grid.querySelectorAll('.pagination-link, .pagination-previous, .pagination-next');
        paginationLinks.forEach((link) => {
            link.addEventListener('click', (event) => {
                event.preventDefault();
                const page = event.target.getAttribute('data-page');
                this.page = parseInt(page);
                this.grid.load();
            });
        });
    };
};