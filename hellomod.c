// hello driver for linux
#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/init.h>
static int __init lkp_init(void)
{
	printk("<1> hello mod from kernel space...\n");
	return 0;
}
static void __exit lkp_cleanup(void)
{
	printk("<1> goodbye mod, leaving the kernel space ...\n");
}
module_init(lkp_init);
module_exit(lkp_cleanup);
MODULE_LICENSE("GPL");
