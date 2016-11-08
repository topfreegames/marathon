const Sequelize = require('sequelize')

module.exports = sequelize => (
  sequelize.define('app', {
    id: {
      type: Sequelize.UUID,
      primaryKey: true,
      defaultValue: Sequelize.UUIDV4,
    },
    key: {
      type: Sequelize.STRING,
      allowNull: false,
      validate: { len: [1, 255] },
    },
    bundleId: {
      type: Sequelize.STRING,
      allowNull: false,
      validate: { len: [1, 2000] },
    },
    createdBy: {
      type: Sequelize.STRING,
      allowNull: false,
      validate: { len: [1, 2000] },
    },
  }, {
    timestamps: true,
    underscored: true,
    indexes: [
      { fields: ['bundleId'], unique: true },
    ],
  })
)
