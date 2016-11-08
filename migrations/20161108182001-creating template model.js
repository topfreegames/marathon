// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

module.exports = {
  up: (queryInterface, Sequelize) => {
    const App = queryInterface.createTable('templates', {
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
      createdAt: {
        type: Sequelize.DATE,
        defaultValue: Sequelize.fn('now'),
        field: 'created_at',
      },
      updatedAt: {
        type: Sequelize.DATE,
        defaultValue: Sequelize.fn('now'),
        field: 'updated_at',
      },
      appId: {
        type: Sequelize.UUID,
        references: {
          model: 'apps',
          key: 'id',
        },
      },
    }).then(() =>
      queryInterface.addIndex('templates', ['appId', 'name', 'locale'], { indicesType: 'UNIQUE' }))

    return App
  },

  down: queryInterface =>
    queryInterface.removeIndex('templates', ['appId', 'name', 'locale']).then(
      () => queryInterface.dropTable('templates')
    ),
}
