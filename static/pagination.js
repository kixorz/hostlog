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
                const target = event.target;
                const disabled = target.hasAttribute('disabled');
                if(disabled) {
                    return;
                }
                const page = target.getAttribute('data-page');
                this.page = parseInt(page, 10);
                this.grid.load();
            });
        });
    };
};