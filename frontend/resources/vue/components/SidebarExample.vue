<template>
  <div class="sidebar-example">
    <h3>Sidebar Management Example</h3>
    
    <div class="controls">
      <button @click="addCustomNavLink" class="btn btn-primary">
        Add Custom Nav Link
      </button>
      
      <button @click="addCustomSupplementalLink" class="btn btn-secondary">
        Add Custom Supplemental Link
      </button>
      
      <button @click="removeCustomLinks" class="btn btn-danger">
        Remove Custom Links
      </button>
      
      <button @click="toggleSidebar" class="btn btn-info">
        Toggle Sidebar
      </button>
    </div>
    
    <div class="info">
      <p><strong>Available Sidebar Methods:</strong></p>
      <ul>
        <li><code>window.sidebar.addNavigationLink(link)</code> - Add navigation link</li>
        <li><code>window.sidebar.addSupplementalLink(link)</code> - Add supplemental link</li>
        <li><code>window.sidebar.removeNavigationLink(id)</code> - Remove navigation link</li>
        <li><code>window.sidebar.removeSupplementalLink(id)</code> - Remove supplemental link</li>
        <li><code>window.sidebar.toggle()</code> - Toggle sidebar visibility</li>
        <li><code>window.sidebar.open()</code> - Open sidebar</li>
        <li><code>window.sidebar.close()</code> - Close sidebar</li>
      </ul>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SidebarExample',
  data() {
    return {
      customLinkId: 1
    }
  },
  methods: {
    addCustomNavLink() {
      if (window.sidebar) {
        window.sidebar.addNavigationLink({
          id: `custom-nav-${this.customLinkId}`,
          title: `Custom Nav ${this.customLinkId}`,
          path: `/custom-${this.customLinkId}`,
          icon: '🔗'
        })
        this.customLinkId++
      } else {
        console.warn('Sidebar not available')
      }
    },
    
    addCustomSupplementalLink() {
      if (window.sidebar) {
        window.sidebar.addSupplementalLink({
          id: `custom-supplemental-${this.customLinkId}`,
          title: `Custom Link ${this.customLinkId}`,
          url: `https://example.com/custom-${this.customLinkId}`,
          target: '_blank',
          icon: '🔗'
        })
        this.customLinkId++
      } else {
        console.warn('Sidebar not available')
      }
    },
    
    removeCustomLinks() {
      if (window.sidebar) {
        // Remove all custom links
        for (let i = 1; i < this.customLinkId; i++) {
          window.sidebar.removeNavigationLink(`custom-nav-${i}`)
          window.sidebar.removeSupplementalLink(`custom-supplemental-${i}`)
        }
        this.customLinkId = 1
      } else {
        console.warn('Sidebar not available')
      }
    },
    
    toggleSidebar() {
      if (window.sidebar) {
        window.sidebar.toggle()
      } else {
        console.warn('Sidebar not available')
      }
    }
  }
}
</script>

<style scoped>
.sidebar-example {
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
  margin: 1rem 0;
}

.controls {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-primary:hover {
  background: #0056b3;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #545b62;
}

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-danger:hover {
  background: #c82333;
}

.btn-info {
  background: #17a2b8;
  color: white;
}

.btn-info:hover {
  background: #138496;
}

.info {
  background: white;
  padding: 1rem;
  border-radius: 4px;
  border-left: 4px solid #007bff;
}

.info ul {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
}

.info code {
  background: #f1f3f4;
  padding: 0.125rem 0.25rem;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.875rem;
}
</style> 