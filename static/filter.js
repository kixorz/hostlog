const filterState = {
    hosts: new Set()
};

const toggleFilter = function(value) {
    if(filterState.hosts.has(value)) {
        filterState.hosts.delete(value);
    } else {
        filterState.hosts.add(value);
    }
};

const registerControl = function(control) {
    control.addEventListener('click', function(event) {
        event.preventDefault();
        const value = event.target.getAttribute('data-host');
        toggleFilter(value);
    });
};

const register = function() {
    const controls = document.getElementsByClassName('host-filter-item');
    for(var i = 0; i < controls.length; ++i) {
        const control = controls[i];
        registerControl(control);
    }
};

document.addEventListener('DOMContentLoaded', register);