module.exports = {
  up: (queryInterface, Sequelize) => {
    const App = queryInterface.createTable('apps', {
      id: {
        type: Sequelize.INTEGER,
        primaryKey: true,
        autoIncrement: true,
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
    })

    return App
  },

  down: queryInterface =>
    queryInterface.dropTable('apps'),
}
