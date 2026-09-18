Looks the featuer and code already running well, and it's good lah. here is the adjustment i need:
1. adding new user with 1 user 1 role, 1 kitchen, 1 cashier, 1 warehouse. admin already exist. 
2. make the information more good on the page
    - on kitchen page looks still showing ProductID and price only and qty only, the product name not correctly showed, i think better for show the sku-code and name instead of ProductID
3. make simple access menu per role, so the detail is like this
    - will be any 2 tables:
        1. access_menus: will be storing the master of access menu, like POS, KDS, Warehouse, Finance, Analytics, Receipt Scanner
        2. access_menus_per_roles: will be storing the config per role can access which access_menu (dynamic)
    - so on Backend will be create new migrations files, and also initiation the seeder
    - after that we need 1 menu on frontend that only accessed by Admin for handle this dynamic access menu, we can call it Setup Role lah.
    - so we need FE and also BE for manage this dynamic access-menu
4. after all already done, i need to implenent logger on Backend side, we can using zerolog i think, for make the Backend is have observablity, and tracable.
    - i think we can have 3 log levels, info, error, and debug
    - if can we can check the log using something like grafana or loki or ELK stack, for make it easy for read and analyze. 
    - i think we will need new configuration on Backend for handle it