// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

import Sequelize from 'sequelize'

module.exports = sequelize => (
  sequelize.define('templates', {
    id: {
      type: Sequelize.UUID,
      primaryKey: true,
      defaultValue: Sequelize.UUIDV4,
    },
    name: {
      type: Sequelize.STRING,
      allowNull: false,
      validate: { len: [1, 255] },
    },
    locale: {
      type: Sequelize.STRING,
      allowNull: false,
      validate: { len: [1, 10] },
      defaultValue: 'en',
    },
    defaults: {
      type: Sequelize.JSONB,
      allowNull: false,
    },
    body: {
      type: Sequelize.JSONB,
      allowNull: false,
    },
    compiledBody: {
      type: Sequelize.STRING,
      allowNull: false,
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
      { fields: ['appId', 'name', 'locale'], unique: true },
    ],
    classMethods: {
      associate: (models) => {
        models.Template.belongsTo(models.App, {
          foreignKey: {
            allowNull: false,
            field: 'appId',
            fieldName: 'appId',
          },
        })
      },
    },
  })
)
