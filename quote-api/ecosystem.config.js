module.exports = {
  apps: [{
    name: 'quote-api',
    script: './index.js',
    max_memory_restart: '1000M',
    instances: 3,
    exec_mode: 'cluster',
    watch: true,
    ignore_watch: ['node_modules', 'assets'],
    env: {
      NODE_ENV: 'development'
    },
    env_production: {
      NODE_ENV: 'production'
    }
  }]
}
