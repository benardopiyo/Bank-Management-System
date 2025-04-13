package main

import (
	"fmt"
	"os"
	"time"
)

var (
	name             string
	pin              int
	availableBalance int = 2000
)

func main() {
	DepositMoney(0)
	MakeWithdrawal(0)
	BalanceInquiry()
}

func DepositMoney(amount int) {
	fmt.Println("Enter Pin:")
	fmt.Scanln(&pin)

	if pin != 3030 {
		fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
		os.Exit(1)
	}
	fmt.Println()

	fmt.Println("Kindly enter your surname:", name)
	// var name string
	fmt.Scanln(&name)
	fmt.Println()

	fmt.Printf("Hello %s! Welcome to KCB Bank. Kindly select Option:", name)
	fmt.Println()
	fmt.Println("1) Balance Inquiry")
	fmt.Println("2) Deposit Money")
	fmt.Println("3) Withdraw Cash")
	fmt.Println("4) Send Money")
	fmt.Println("5) Loans & Savings")
	fmt.Println("6) My Account")
	fmt.Println()

	fmt.Println("Select Option:")
	var option int
	fmt.Scanln(&option)
	fmt.Println()

	if option == 1 || option == 2 || option == 3 || option == 4 || option == 5 || option == 6 {
		fmt.Printf("Dear %s, you have selected Option %d", name, option)
	} else {
		fmt.Println("Error: kindly provide a digit between 1-6")
		os.Exit(1)
	}
	// fmt.Println()

	if option == 1 {
		fmt.Println()
		fmt.Printf("For Balance Inquiry, kindly enter Pin:")
		fmt.Println()

		// var pin int
		fmt.Scanln(&pin)

		if pin != 3030 {
			fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
			os.Exit(1)
		}
		fmt.Println()
		fmt.Print("Bank Actual Balance as at ", time.Now())
		fmt.Println(" is KES", availableBalance)

		// to proceed?
		fmt.Println()
		fmt.Printf("Dear %s, would you like to proceed?", name)
		fmt.Println()
		fmt.Println("1) Yes")
		fmt.Println("2) No")
		fmt.Println()

		fmt.Println("Select Choice:")
		var makeChoice int
		fmt.Scanln(&makeChoice)

		if makeChoice == 1 {
			fmt.Printf("Hello %s! Welcome to KCB Bank. Kindly select Option:", name)
			fmt.Println()
			fmt.Println("1) Balance Inquiry")
			fmt.Println("2) Deposit Money")
			fmt.Println("3) Withdraw Cash")
			fmt.Println("4) Send Money")
			fmt.Println("5) Loans & Savings")
			fmt.Println("6) My Account")
			fmt.Println()

			fmt.Println("Select Option:")
			var option int
			fmt.Scanln(&option)

			if option == 2 {
				// make deposit
				fmt.Println()
				fmt.Printf("Dear %s, you have selected Option %d", name, option)
				fmt.Println()

				fmt.Println()
				fmt.Println("For Money Deposit, kindly enter Deposit Amount:")
				fmt.Scanln(&amount)
				availableBalance += amount

				fmt.Println()
				fmt.Println("For Deposit, kindly enter Pin:")

				// var pin int
				fmt.Scanln(&pin)

				if pin != 3030 {
					fmt.Println("Incorrect pin: re-enter the correct pin, 2 more trials") // to add 2 trials
					os.Exit(1)
				}

				fmt.Println()
				fmt.Printf("Dear %s, Deposit of KES %d was successful: Actual Balance is KES: %d", name, amount, availableBalance) // include time.Now()
				fmt.Println()
				// fmt.Println("ACTUAL BALANCE:", availableBalance)
				fmt.Println()
				fmt.Printf("Dear %s, would you like to proceed?", name)
				fmt.Println()
				fmt.Println("1) Yes")
				fmt.Println("2) No")

				fmt.Println()
				fmt.Println("Select Choice:")
				var makeChoice int
				fmt.Scanln(&makeChoice)

				fmt.Println()
				if makeChoice == 1 {
					fmt.Printf("Hello %s! Welcome to KCB Bank. Kindly select Option:", name)
					fmt.Println()
					fmt.Println("1) Balance Inquiry")
					fmt.Println("2) Deposit Money")
					fmt.Println("3) Withdraw Cash")
					fmt.Println("4) Send Money")
					fmt.Println("5) Loans & Savings")
					fmt.Println("6) My Account")
					fmt.Println()

					fmt.Println("Select Option:")
					var option int
					fmt.Scanln(&option)

					if option == 3 {
						fmt.Println()
						fmt.Printf("Dear %s, you have selected Option %d", name, option)
						fmt.Println()

						fmt.Printf("For Cash Withdrawal, kindly enter Withdrawal Amount:")
						fmt.Scanln(&amount)

						fmt.Println()
						fmt.Println("For Cash Withdrawal, kindly enter Pin:")
						fmt.Scanln(&pin)
						fmt.Println()

						if pin != 3030 {
							fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
							os.Exit(1)
						}

						if availableBalance >= amount {
							availableBalance = availableBalance - amount
							fmt.Printf("Dear %s, Withdrawal of KES %d was successful: Actual Balance is KES %d:", name, amount, availableBalance)
							fmt.Println(" Withrawal successfully done as at", time.Now())
						} else {
							fmt.Printf("Dear %s, You have limited funds in your account, kindly top up to continue enjoying the services", name)
						}
						fmt.Println()
					}
				} else {
					os.Exit(1)
				}

				fmt.Println()
				fmt.Printf("Dear %s, would you like to proceed?", name)
				fmt.Println()
				fmt.Println("1) Yes")
				fmt.Println("2) No")

				fmt.Println()
				fmt.Println("Select Choice:")
				// var makeChoice int
				fmt.Scanln(&makeChoice)

				fmt.Println()
				if makeChoice == 1 {
					fmt.Printf("Hello %s! Welcome to KCB Bank. Kindly select Option:", name)
					fmt.Println()
					fmt.Println("1) Balance Inquiry")
					fmt.Println("2) Deposit Money")
					fmt.Println("3) Withdraw Cash")
					fmt.Println("4) Send Money")
					fmt.Println("5) Loans & Savings")
					fmt.Println("6) My Account")
					fmt.Println()

					fmt.Println("Select Option:")
					var option int
					fmt.Scanln(&option)

					if option == 4 {
						fmt.Println()
						fmt.Printf("Dear %s, you have selected Option %d", name, option)
						fmt.Println()

						fmt.Printf("For Send Money, kindly select Beneficiary:")
						fmt.Println()
						fmt.Println("1) MPESA")
						fmt.Println("2) AIRTEL")
						fmt.Println()

						fmt.Println("Beneficiary:")
						var Beneficiary int
						fmt.Scanln(&Beneficiary)

						if Beneficiary == 1 {
							fmt.Println("Kindly enter MPESA Number, 07.. or 01..:")
							fmt.Scanln(&Beneficiary)
							fmt.Println()
							fmt.Println("Kindly enter Amount:")
							fmt.Scanln(&amount)
							fmt.Println()
							fmt.Println("For Send Money, kindly enter Pin:")
							fmt.Scanln(&pin)
							fmt.Println()

							if pin != 3030 {
								fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
								os.Exit(1)
							}

							if availableBalance >= amount {
								availableBalance = availableBalance - amount
								fmt.Printf("Dear %s, KES %d was successfully sent to NAME OF BENEFICIARY of 0710000000: Actual Balance is KES %d:", name, amount, availableBalance)
							} else {
								fmt.Printf("Dear %s, You have limited funds in your account, kindly top up to continue enjoying the services", name)
							}
							fmt.Println()

						} else if Beneficiary == 2 {
							fmt.Println("Kindly enter AIRTEL Number, 07..:")
							fmt.Scanln(&Beneficiary)
							fmt.Println()
							fmt.Println("Kindly enter Amount:")
							fmt.Scanln(&amount)
							fmt.Println()
							fmt.Println("For Send Money, kindly enter Pin:")
							fmt.Scanln(&pin)
							fmt.Println()

							if pin != 3030 {
								fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
								os.Exit(1)
							}

							if availableBalance >= amount {
								availableBalance = availableBalance - amount
								fmt.Printf("Dear %s, KES %d was successfully sent to NAME OF BENEFICIARY of 07...: Actual Balance is KES %d:", name, amount, availableBalance)
							} else {
								fmt.Printf("Dear %s, You have limited funds in your account, kindly top up to continue enjoying the services", name)
							}
							fmt.Println()
						}

					}
				}
			}

		} else {
			os.Exit(1)
		}
	} else if option == 2 {
		// make deposit
		fmt.Println()
		fmt.Println("Enter Deposit Amount:")
		fmt.Scanln(&amount)
		availableBalance += amount

		fmt.Println()
		fmt.Println("Enter Pin:")

		// var pin int
		fmt.Scanln(&pin)

		if pin != 3030 {
			fmt.Println("Incorrect pin: re-enter the correct pin, 2 more trials") // to add 2 trials
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("Dear %s, Deposit of KES %d was successful: Actual Balance is KES: %d", name, amount, availableBalance) // include time.Now()
		fmt.Println()
		fmt.Println("ACTUAL BALANCE:", availableBalance)
		fmt.Println()

		fmt.Printf("Dear %s, would you like to proceed?", name)
		fmt.Println()
		fmt.Println("1) Yes")
		fmt.Println("2) No")

		fmt.Println("Select Choice:")
		var makeChoice int
		fmt.Scanln(&makeChoice)

		if makeChoice == 1 {
			fmt.Printf("Hello %s! Welcome to KCB Bank. Kindly select Option:", name)
			fmt.Println()
			fmt.Println("1) Balance Inquiry")
			fmt.Println("2) Deposit Money")
			fmt.Println("3) Withdraw Cash")
			fmt.Println("4) Send Money")
			fmt.Println("5) Loans & Savings")
			fmt.Println("6) My Account")
			fmt.Println()

			fmt.Println("Select Option:")
			var option int
			fmt.Scanln(&option)
		} else {
			os.Exit(1)
		}
	} else if option == 3 {
		fmt.Println(3)
	} else if option == 4 {
		fmt.Println(4)
	} else if option == 5 {
		fmt.Println(5)
	} else {
		fmt.Println(6)
	}

	// // make deposit
	// fmt.Println("Enter deposit amount:")
	// fmt.Scanln(&amount)
	// availableBalance += amount

	// fmt.Println("Enter Pin:")

	// // var pin int
	// fmt.Scanln(&pin)

	// if pin != 3030 {
	// 	fmt.Println("Incorrect pin: re-enter the correct pin, 2 more trials") // to add 2 trials
	// 	os.Exit(1)
	// }

	// fmt.Printf("Dear %s, Deposit of KES %d was successful: Actual Balance is KES: %d", name, amount, availableBalance) // include time.Now()
	// fmt.Println()
	// fmt.Println("ACTUAL BALANCE:", availableBalance)
	// fmt.Println()
}

func MakeWithdrawal(amount int) {
	// fmt.Scanln(&name)
	// fmt.Printf("Hello %s, to make withdrawal, kindly enter withdrawal amout:", name)
	// fmt.Scanln(&amount)

	// fmt.Println("Enter Pin:")
	// fmt.Scanln(&pin)

	// if pin != 3030 {
	// 	fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
	// 	os.Exit(1)
	// }

	// if availableBalance >= amount {
	// 	availableBalance = availableBalance - amount
	// 	fmt.Printf("Dear %s, withdrawal of KES %d was successful: Actual Balance is KES %d:", name, amount, availableBalance)
	// 	fmt.Println(" Withrawal successfully done as at:", time.Now())
	// } else {
	// 	fmt.Printf("Dear %s, You have limited funds in your account, kindly top up to continue enjoying the services", name)
	// }
	// fmt.Println()
}

func BalanceInquiry() {
	fmt.Printf("Hello %s, to check Bank balance, kindly enter pin:", name)

	// var pin int
	fmt.Scanln(&pin)

	if pin != 3030 {
		fmt.Println("Incorrect pin: you have two more attempts") // to add 2 more trials
		os.Exit(1)
	}
	fmt.Print("Bank Actual Balance as at: ", time.Now())
	fmt.Println(" is KES:", availableBalance)
}
