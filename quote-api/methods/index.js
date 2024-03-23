const crypto = require('crypto')

const generate = require('./generate')

const methods = {
  generate
}

module.exports = async (method, parm) => {
  if (methods[method]) {
    return await methods[method](parm)
  } else {
    return {
      error: 'method not found'
    }
  }
}
