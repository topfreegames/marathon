// marathon
// https://github.com/topfreegames/marathon
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

module.exports = {
  up: (queryInterface, Sequelize) => {
    const Job = queryInterface.createTable('jobs', {
      id: {
        type: Sequelize.UUID,
        primaryKey: true,
        defaultValue: Sequelize.UUIDV4,
      },
      totalBatches: {
        type: Sequelize.INTEGER,
        validate: { min: 1 },
        field: 'total_batches',
      },
      completedBatches: {
        type: Sequelize.INTEGER,
        allowNull: false,
        validate: { min: 0 },
        defaultValue: 0,
        field: 'completed_batches',
      },
      completedAt: {
        type: Sequelize.DATE,
        defaultValue: Sequelize.fn('now'),
        field: 'completed_at',
      },
      filters: {
        type: Sequelize.JSONB,
      },
      csvUrl: {
        type: Sequelize.STRING,
        validate: { isUrl: true },
        field: 'csv_url',
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
      appId: {
        type: Sequelize.UUID,
        field: 'app_id',
        references: {
          model: 'apps',
          key: 'id',
        },
      },
      templateId: {
        type: Sequelize.UUID,
        field: 'template_id',
        references: {
          model: 'templates',
          key: 'id',
        },
      },
    })

    return Job
  },

  down: queryInterface => () => queryInterface.dropTable('templates'),
}
