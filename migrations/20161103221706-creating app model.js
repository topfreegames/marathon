// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

module.exports = {
  up: (queryInterface, Sequelize) => {
    const App = queryInterface.createTable('apps', {
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
        field: 'bundle_id',
      },
      createdBy: {
        type: Sequelize.STRING,
        allowNull: false,
        validate: { len: [1, 2000] },
        field: 'created_by',
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
    }).then(() => queryInterface.addIndex('apps', ['bundle_id'], { indicesType: 'UNIQUE' }))

    return App
  },

  down: queryInterface =>
    queryInterface.removeIndex('apps', ['bundle_id']).then(() => queryInterface.dropTable('apps')),
}
