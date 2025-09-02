<template>
  <div class="not-found-view">
    <div class="not-found-container">
      <div class="not-found-content">
        <h1>404</h1>
        <h2>Page Not Found</h2>
        
        <div class="actions">
          <button class = "button good" @click="goToHome">
            Go to Home
          </button>
          <button class="button neutral" @click="goBack">
            Go Back
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NotFoundView',
  methods: {
    goBack() {
      this.$router.go(-1)
    },
    goToHome() {
      this.$router.push('/')
    }
  }
}
</script>

<style scoped>

.not-found-content {
  padding: 3rem 2rem;
  text-align: center;
}

.not-found-content h1 {
  font-size: 6rem;
  margin: 0;
  font-weight: 700;
  line-height: 1;
}

.not-found-content h2 {
  font-size: 2rem;
  margin: 0 0 1rem 0;
  color: #333;
}

.not-found-content p {
  font-size: 1.1rem;
  color: #666;
  margin-bottom: 2rem;
}
</style> 