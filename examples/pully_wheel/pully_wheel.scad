

	include <BOSL2/std.scad>;
    
    
    $fn = 100;

outerRadius = 7.5;
innerRadius = 5.55;
pullyLength = outerRadius*2;
pullyChamfer =1.5;
    
module pully_wheel(){
difference(){
	cyl(r=outerRadius, h=pullyLength, center=true, chamfer=-pullyChamfer);
    
	cyl(r=innerRadius, h=pullyLength+1, center=true, chamfer=-pullyChamfer);
    }
}

pully_wheel();
