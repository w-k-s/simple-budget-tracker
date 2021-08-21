package test

// - test update category last used
// - given a transfer with category of different user, when record is saved, then error is returned
// - given a transfer sent with positive amount, when record is saved, sender account amount is negative
// - given a transfer sent with a negative amount, when record is saved, sender account amount is negative
// - given a transfer sent with positive amount, when record is saved, receiver account amount is positive
// - given a transfer sent with negative amount, when record is saved, receiver account amount is positive
// - test get account by id (should show correct total balance)
// - given an account receives transfer, when receiver account edits transfer, then error is returned
// - given an account receives transfer, when receiver account deletes transfer, then error is returned
