// Populate the sidebar
//
// This is a script, and not included directly in the page, to control the total size of the book.
// The TOC contains an entry for each page, so if each page includes a copy of the TOC,
// the total size of the page becomes O(n**2).
class MDBookSidebarScrollbox extends HTMLElement {
    constructor() {
        super();
    }
    connectedCallback() {
        this.innerHTML = '<ol class="chapter"><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="index.html">autotidy</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="quick-start.html">quick start</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="installation/index.html">installation</a></span><ol class="section"><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="installation/linux.html">linux</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="installation/macos.html">macos</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="installation/windows.html">windows</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="installation/nix.html">nix</a></span></li></ol><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="configuration.html">configuration</a></span><ol class="section"><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="rules.html">rules</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/index.html">filters</a></span><ol class="section"><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/name.html">name</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/extension.html">extension</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/file_size.html">file_size</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/file-type.html">file_type</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/date-modified.html">date_modified</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/date-accessed.html">date_accessed</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/date-created.html">date_created</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/date-changed.html">date_changed</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="filters/mime-type.html">mime_type</a></span></li></ol><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/index.html">actions</a></span><ol class="section"><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/move.html">move</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/copy.html">copy</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/rename.html">rename</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/delete.html">delete</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/trash.html">trash</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="actions/log.html">log</a></span></li></ol><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="templates.html">templates</a></span></li><li class="chapter-item expanded "><span class="chapter-link-wrapper"><a href="options.html">additional options</a></span></li></ol></li></ol>';
        // Set the current, active page, and reveal it if it's hidden
        let current_page = document.location.href.toString().split('#')[0].split('?')[0];
        if (current_page.endsWith('/')) {
            current_page += 'index.html';
        }
        const links = Array.prototype.slice.call(this.querySelectorAll('a'));
        const l = links.length;
        for (let i = 0; i < l; ++i) {
            const link = links[i];
            const href = link.getAttribute('href');
            if (href && !href.startsWith('#') && !/^(?:[a-z+]+:)?\/\//.test(href)) {
                link.href = path_to_root + href;
            }
            // The 'index' page is supposed to alias the first chapter in the book.
            if (link.href === current_page
                || i === 0
                && path_to_root === ''
                && current_page.endsWith('/index.html')) {
                link.classList.add('active');
                let parent = link.parentElement;
                while (parent) {
                    if (parent.tagName === 'LI' && parent.classList.contains('chapter-item')) {
                        parent.classList.add('expanded');
                    }
                    parent = parent.parentElement;
                }
            }
        }
        // Track and set sidebar scroll position
        this.addEventListener('click', e => {
            if (e.target.tagName === 'A') {
                const clientRect = e.target.getBoundingClientRect();
                const sidebarRect = this.getBoundingClientRect();
                sessionStorage.setItem('sidebar-scroll-offset', clientRect.top - sidebarRect.top);
            }
        }, { passive: true });
        const sidebarScrollOffset = sessionStorage.getItem('sidebar-scroll-offset');
        sessionStorage.removeItem('sidebar-scroll-offset');
        if (sidebarScrollOffset !== null) {
            // preserve sidebar scroll position when navigating via links within sidebar
            const activeSection = this.querySelector('.active');
            if (activeSection) {
                const clientRect = activeSection.getBoundingClientRect();
                const sidebarRect = this.getBoundingClientRect();
                const currentOffset = clientRect.top - sidebarRect.top;
                this.scrollTop += currentOffset - parseFloat(sidebarScrollOffset);
            }
        } else {
            // scroll sidebar to current active section when navigating via
            // 'next/previous chapter' buttons
            const activeSection = document.querySelector('#mdbook-sidebar .active');
            if (activeSection) {
                activeSection.scrollIntoView({ block: 'center' });
            }
        }
        // Toggle buttons
        const sidebarAnchorToggles = document.querySelectorAll('.chapter-fold-toggle');
        function toggleSection(ev) {
            ev.currentTarget.parentElement.parentElement.classList.toggle('expanded');
        }
        Array.from(sidebarAnchorToggles).forEach(el => {
            el.addEventListener('click', toggleSection);
        });
    }
}
window.customElements.define('mdbook-sidebar-scrollbox', MDBookSidebarScrollbox);

