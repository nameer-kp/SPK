#!/usr/bin/env python3
"""
Interactive CLI Booking Tool for iTravel API
Allows users to configure booking parameters and generate booking requests
"""

import json
import os
from datetime import datetime, timedelta
from typing import Optional


# Available options based on request.json analysis
AVAILABLE_MARKETS = ["US", "UK", "AU", "CA", "DE", "FR", "IT", "ES"]
AVAILABLE_CURRENCIES = ["USD", "EUR", "GBP", "AUD", "CAD"]
BOOKING_TYPES = ["IND", "GRP"]  # Individual, Group
PRODUCT_TYPES = ["CRU", "PKG"]  # Cruise, Package
PAX_TYPES = {"A": "Adult", "C": "Child", "I": "Infant"}

# API Base URL
BASE_URL = "https://sandbox.itravel.ibsplc.org/iTravel"


class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    CYAN = '\033[96m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


def colored(text: str, color: str) -> str:
    """Return colored text for terminal output"""
    return f"{color}{text}{Colors.ENDC}"


def print_header(text: str):
    """Print a styled header"""
    print("\n" + "=" * 60)
    print(colored(f"  {text}", Colors.HEADER + Colors.BOLD))
    print("=" * 60)


def print_section(text: str):
    """Print a section header"""
    print("\n" + colored(f"--- {text} ---", Colors.CYAN))


def print_success(text: str):
    """Print success message"""
    print(colored(f"[OK] {text}", Colors.GREEN))


def print_error(text: str):
    """Print error message"""
    print(colored(f"[ERROR] {text}", Colors.RED))


def print_info(text: str):
    """Print info message"""
    print(colored(f"[INFO] {text}", Colors.YELLOW))


def get_input(prompt: str, default: Optional[str] = None, required: bool = True) -> str:
    """Get user input with optional default value"""
    if default:
        display_prompt = f"{prompt} [{colored(default, Colors.YELLOW)}]: "
    else:
        display_prompt = f"{prompt}: "

    while True:
        value = input(display_prompt).strip()
        if not value and default:
            return default
        if not value and required:
            print_error("This field is required. Please enter a value.")
            continue
        if not value and not required:
            return ""
        return value


def get_choice(prompt: str, choices: list, default: Optional[str] = None) -> str:
    """Get user choice from a list of options"""
    print(f"\n{prompt}")
    for i, choice in enumerate(choices, 1):
        marker = " *" if choice == default else ""
        print(f"  {colored(str(i), Colors.CYAN)}. {choice}{marker}")

    while True:
        selection = input(f"Enter choice (1-{len(choices)}) [{default or '1'}]: ").strip()
        if not selection and default:
            return default
        if not selection:
            return choices[0]
        try:
            idx = int(selection) - 1
            if 0 <= idx < len(choices):
                return choices[idx]
            print_error(f"Please enter a number between 1 and {len(choices)}")
        except ValueError:
            # Check if they entered the choice directly
            if selection.upper() in [c.upper() for c in choices]:
                return selection.upper()
            print_error("Invalid input. Please enter a number or valid choice.")


def get_number(prompt: str, default: int = 1, min_val: int = 1, max_val: int = 100) -> int:
    """Get a number input with validation"""
    while True:
        value = input(f"{prompt} [{colored(str(default), Colors.YELLOW)}]: ").strip()
        if not value:
            return default
        try:
            num = int(value)
            if min_val <= num <= max_val:
                return num
            print_error(f"Please enter a number between {min_val} and {max_val}")
        except ValueError:
            print_error("Please enter a valid number")


def get_date(prompt: str, default: Optional[str] = None) -> str:
    """Get date input in YYYY-MM-DD format"""
    if not default:
        default = (datetime.now() + timedelta(days=30)).strftime("%Y-%m-%d")

    while True:
        value = input(f"{prompt} (YYYY-MM-DD) [{colored(default, Colors.YELLOW)}]: ").strip()
        if not value:
            return default
        try:
            datetime.strptime(value, "%Y-%m-%d")
            return value
        except ValueError:
            print_error("Invalid date format. Please use YYYY-MM-DD")


def get_yes_no(prompt: str, default: bool = True) -> bool:
    """Get yes/no input"""
    default_str = "Y/n" if default else "y/N"
    while True:
        value = input(f"{prompt} [{default_str}]: ").strip().lower()
        if not value:
            return default
        if value in ['y', 'yes']:
            return True
        if value in ['n', 'no']:
            return False
        print_error("Please enter 'y' or 'n'")


def collect_passenger_details(pax_num: int, pax_type: str, is_lead: bool = False) -> dict:
    """Collect details for a single passenger"""
    print(f"\n  Passenger {pax_num} ({PAX_TYPES.get(pax_type, 'Adult')}){'  [LEAD]' if is_lead else ''}")
    print("  " + "-" * 40)

    passenger = {
        "RefId": pax_num,
        "PaxType": pax_type,
        "IsLeadPassenger": is_lead,
        "paxRefId": pax_num,
        "IsSeated": True,
        "LoyaltyTier": "",
        "LoyaltyTierValue": ""
    }

    if is_lead:
        passenger["FirstName"] = get_input("    First Name", "Guest")
        passenger["LastName"] = get_input("    Last Name", "Traveler")
        passenger["memberContactInformation"] = [{"addressType": "H"}]
        passenger["Contacts"] = {
            "Contact": [{
                "Email": get_input("    Email", "", required=False),
                "ContactType": "H",
                "Address": {
                    "AddressLineOne": "",
                    "City": "",
                    "CountryCode": "",
                    "StateProv": "",
                    "ZipCode": "",
                    "CountryName": "",
                    "IsVerifiedAddress": False
                }
            }]
        }

    return passenger


def create_guest_details(passengers: list) -> dict:
    """Create guest details structure for cruise search"""
    guests = []
    for pax in passengers:
        guests.append({
            "CabinSequenceNumber": 1,
            "PaxType": "Adult" if pax["PaxType"] == "A" else ("Child" if pax["PaxType"] == "C" else "Infant"),
            "PassengerreferenceID": pax["RefId"]
        })
    return {"Guest": guests}


def create_customer_counts(passengers: list) -> list:
    """Create customer counts structure for package search"""
    customer_counts = []
    for pax in passengers:
        customer_counts.append({
            "Code": pax["PaxType"],
            "PaxRefs": pax["RefId"],
            "PassengerreferenceID": pax["RefId"],
            "ExtraBedsReqd": "0",
            "Quantity": "1",
            "Age": None,
            "BirthDate": "",
            "LeadPax": pax.get("IsLeadPassenger", False),
            "LoyaltyTier": "",
            "PointsToNextTier": "",
            "PersonType": ""
        })
    return [{
        "Room": 0,
        "CabinSeqNumber": "1",
        "CustomerCount": customer_counts
    }]


def generate_cruise_availability_request(config: dict) -> dict:
    """Generate cruise availability search request"""
    return {
        "url": f"{BASE_URL}/selling-availability/api/public-availability/v1/rest/avl/availability/cruiseAggrAvailabilitySearch",
        "method": "POST",
        "body": {
            "ProviderCode": "",
            "PriceCurrency": config["currency"],
            "Market": config["market"],
            "Data": {
                "IBS_CruAggregateRQ": {
                    "CruiseSearch": {
                        "Itinerary": config.get("itinerary", ""),
                        "Departure": {
                            "Start": config["departure_start"],
                            "End": config["departure_end"]
                        },
                        "OverridePassengerControlCapacity": "No",
                        "BedConfiguration": ""
                    },
                    "AdhocCriteria": {
                        "SendSailingInfo": True,
                        "SendDepartureDetails": True,
                        "SendShipDetails": config.get("include_ship_details", False),
                        "SendDeckDetails": False,
                        "SendDeckPlanDetails": False,
                        "SendDiningRoomDetails": False,
                        "SendPublicRoomDetails": False,
                        "SendShipPolicies": False,
                        "SendCategoryDetails": config.get("include_category_details", False),
                        "SendCabinDetails": config.get("include_cabin_details", False),
                        "SendFare": config.get("include_fare", False),
                        "SendBedConfigDetails": False,
                        "SendFacilityDetails": False,
                        "SendItineraryDetails": True,
                        "SendIndicativeRatesforDeparture": "For Departure",
                        "SendPromotions": "Pre Applied Promotions",
                        "IsBasicSearch": True
                    },
                    "CruiseParticulars": {
                        "BookingType": config["booking_type"],
                        "ServiceOrigin": "iTravelB2B"
                    },
                    "GuestDetails": create_guest_details(config["passengers"]),
                    "Offset": 0,
                    "SortCriteria": 1
                }
            },
            "mData": {
                "AdvancedSearchParameters": {
                    "Parameter": []
                }
            }
        }
    }


def generate_package_availability_request(config: dict) -> dict:
    """Generate package availability search request"""
    return {
        "url": f"{BASE_URL}/selling-availability/api/public-availability/v1/rest/avl/availability/packageAvailabilitySearch",
        "method": "POST",
        "body": {
            "BusinessParams": {
                "BookingType": config["booking_type"],
                "GroupCode": ""
            },
            "ProviderCode": None,
            "Market": config["market"],
            "PriceCurrency": config["currency"],
            "Data": {
                "IBS_PkgSearchRQ": {
                    "MoreDataEchoToken": "string",
                    "MaxResponses": "99",
                    "SortOrder": "A",
                    "CustomerCounts": create_customer_counts(config["passengers"]),
                    "Criteria": {
                        "AvailableOnlyIndicator": "true",
                        "PackageRef": {
                            "ExactMatch": "true",
                            "PackageCodes": config.get("package_code", "")
                        },
                        "DepartDateRange": {
                            "Start": config["departure_start"],
                            "End": config["departure_end"]
                        },
                        "PriceRange": {},
                        "EmbeddedCruise": {
                            "ShipNames": [],
                            "CruiseID": config.get("cruise_id", ""),
                            "CabinSequenceNumber": "1"
                        },
                        "EmbeddedAir": {},
                        "RatePlanCandidate": {
                            "ExactMatch": "true"
                        }
                    },
                    "DeparturesWanted": "true",
                    "DetailsWanted": "false",
                    "PricesWanted": "true"
                }
            }
        }
    }


def generate_shopping_cart_request(config: dict) -> dict:
    """Generate shopping cart create request"""
    lead_pax = next((p for p in config["passengers"] if p.get("IsLeadPassenger")), config["passengers"][0])

    return {
        "url": f"{BASE_URL}/selling/api/public-shoppingcart/v1/rest/crt/shoppingcart/create",
        "method": "POST",
        "body": {
            "userId": config.get("user_id", "user@example.com"),
            "currency": config["currency"],
            "addedDate": datetime.now().isoformat(),
            "emailId": lead_pax.get("Contacts", {}).get("Contact", [{}])[0].get("Email", ""),
            "firstName": lead_pax.get("FirstName", "Guest"),
            "lastName": lead_pax.get("LastName", "Traveler"),
            "CartInfo": {
                "Passengers": {
                    "Passenger": config["passengers"]
                },
                "BookingContact": {
                    "Email": lead_pax.get("Contacts", {}).get("Contact", [{}])[0].get("Email", ""),
                    "FirstName": lead_pax.get("FirstName", "Guest"),
                    "LastName": lead_pax.get("LastName", "Traveler"),
                    "MiddleName": "",
                    "FirstName_En": lead_pax.get("FirstName", "Guest"),
                    "LastName_En": lead_pax.get("LastName", "Traveler"),
                    "MiddleName_En": "",
                    "PhoneNumber": "",
                    "ZipCode": "",
                    "Title": "",
                    "ContactType": "CUSTOMER",
                    "DateOfBirth": "",
                    "CrmID": ""
                }
            },
            "lineitems": {
                "lineitem": []
            }
        }
    }


def display_config_summary(config: dict):
    """Display a summary of the booking configuration"""
    print_section("Configuration Summary")
    print(f"  Market:           {colored(config['market'], Colors.GREEN)}")
    print(f"  Currency:         {colored(config['currency'], Colors.GREEN)}")
    print(f"  Booking Type:     {colored(config['booking_type'], Colors.GREEN)}")
    print(f"  Product Type:     {colored(config['product_type'], Colors.GREEN)}")
    print(f"  Departure Range:  {colored(config['departure_start'], Colors.GREEN)} to {colored(config['departure_end'], Colors.GREEN)}")
    print(f"  Total Passengers: {colored(str(len(config['passengers'])), Colors.GREEN)}")

    adults = sum(1 for p in config['passengers'] if p['PaxType'] == 'A')
    children = sum(1 for p in config['passengers'] if p['PaxType'] == 'C')
    infants = sum(1 for p in config['passengers'] if p['PaxType'] == 'I')

    print(f"    - Adults:       {adults}")
    print(f"    - Children:     {children}")
    print(f"    - Infants:      {infants}")

    if config.get("num_bookings", 1) > 1:
        print(f"  Bookings to Gen:  {colored(str(config['num_bookings']), Colors.GREEN)}")


def export_requests(requests: list, filename: str):
    """Export generated requests to a JSON file"""
    with open(filename, 'w') as f:
        json.dump(requests, f, indent=2)
    print_success(f"Exported {len(requests)} request(s) to {filename}")


def main():
    """Main CLI entry point"""
    print_header("iTravel Interactive Booking CLI Tool")
    print(colored("Configure your booking parameters step by step\n", Colors.CYAN))

    config = {}

    # Basic Configuration
    print_section("Basic Configuration")
    config["market"] = get_choice("Select Market:", AVAILABLE_MARKETS, default="US")
    config["currency"] = get_choice("Select Currency:", AVAILABLE_CURRENCIES, default="USD")
    config["booking_type"] = get_choice("Select Booking Type:", BOOKING_TYPES, default="IND")
    config["product_type"] = get_choice("Select Product Type:", PRODUCT_TYPES, default="CRU")

    # Travel Dates
    print_section("Travel Dates")
    default_start = (datetime.now() + timedelta(days=30)).strftime("%Y-%m-%d")
    default_end = (datetime.now() + timedelta(days=180)).strftime("%Y-%m-%d")
    config["departure_start"] = get_date("Departure Start Date", default_start)
    config["departure_end"] = get_date("Departure End Date", default_end)

    # Passenger Configuration
    print_section("Passenger Configuration")
    num_adults = get_number("Number of Adults", default=2, min_val=1, max_val=10)
    num_children = get_number("Number of Children", default=0, min_val=0, max_val=8)
    num_infants = get_number("Number of Infants", default=0, min_val=0, max_val=4)

    config["passengers"] = []
    pax_num = 1

    # Collect lead passenger details
    print_section("Passenger Details")
    print_info("Enter details for lead passenger (others will use defaults)")

    for i in range(num_adults):
        is_lead = (i == 0)
        config["passengers"].append(collect_passenger_details(pax_num, "A", is_lead))
        pax_num += 1

    for i in range(num_children):
        config["passengers"].append(collect_passenger_details(pax_num, "C", False))
        pax_num += 1

    for i in range(num_infants):
        config["passengers"].append(collect_passenger_details(pax_num, "I", False))
        pax_num += 1

    # Product-specific options
    print_section("Product Options")
    if config["product_type"] == "CRU":
        config["itinerary"] = get_input("Itinerary Code (optional)", "", required=False)
        config["include_ship_details"] = get_yes_no("Include ship details?", default=False)
        config["include_fare"] = get_yes_no("Include fare information?", default=True)
    else:
        config["package_code"] = get_input("Package Code (optional)", "", required=False)
        config["cruise_id"] = get_input("Cruise ID (optional)", "", required=False)

    # Number of bookings
    print_section("Booking Generation")
    config["num_bookings"] = get_number("How many bookings to generate?", default=1, min_val=1, max_val=50)

    # Summary
    display_config_summary(config)

    # Confirm and generate
    print_section("Generate Requests")
    if not get_yes_no("Proceed with generation?", default=True):
        print_info("Operation cancelled.")
        return

    # Generate requests
    requests = []
    for i in range(config["num_bookings"]):
        if config["product_type"] == "CRU":
            requests.append(generate_cruise_availability_request(config))
        else:
            requests.append(generate_package_availability_request(config))

        # Add shopping cart request
        requests.append(generate_shopping_cart_request(config))

    print_success(f"Generated {len(requests)} request(s)")

    # Preview option
    if get_yes_no("Preview generated requests?", default=False):
        print("\n" + json.dumps(requests, indent=2))

    # Export option
    if get_yes_no("Export to JSON file?", default=True):
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        default_filename = f"booking_requests_{timestamp}.json"
        filename = get_input("Output filename", default_filename)
        export_requests(requests, filename)

    print_header("Complete!")
    print(colored("Thank you for using the iTravel Booking CLI Tool\n", Colors.GREEN))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n" + colored("Operation cancelled by user.", Colors.YELLOW))
    except Exception as e:
        print_error(f"An error occurred: {str(e)}")
        raise
